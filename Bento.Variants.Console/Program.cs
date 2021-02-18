using System.Text;
using System.Linq;
using System;
using System.IO;
using System.Collections.Generic;
using System.Collections.Concurrent;
using System.Diagnostics;
using System.Threading.Tasks;

using Bento.Variants.XCC;

using Nest;

namespace Bento.Variants.Console
{
    public class Program
    {
        public static object HttpCallLockObject = new object();
        static void Main(string[] args)
        {
            System.Console.WriteLine("Hello World!");

            string vcfFilesPath = null;
            string url = null;
            string esUsername = null;
            string esPassword = null;

            // Validate arguments
            int argNum = 0;
            foreach (string arg in args)
            {
                if (arg.StartsWith("--"))
                {
                    switch(arg)
                    {
                        case "--vcfPath":
                            if(args.Length >= argNum+1)    
                                vcfFilesPath = $"{System.IO.Directory.GetCurrentDirectory()}/{args[argNum+1]}";
                            break;
                            
                        case "--elasticsearchUrl":
                            if(args.Length >= argNum+1)    
                                url = $"{args[argNum+1]}";
                            break;

                        case "--elasticsearchUsername":
                            if(args.Length >= argNum+1)    
                                esUsername = $"{args[argNum+1]}";
                            break;
                            
                        case "--elasticsearchPassword":
                            if(args.Length >= argNum+1)    
                                esPassword = $"{args[argNum+1]}";
                            break;
                    }
                }

                argNum++;
            }

            if(string.IsNullOrEmpty(vcfFilesPath))
                throw new Exception("Missing --vcfPath argument!");

            if(string.IsNullOrEmpty(url))
                throw new Exception("Missing --elasticsearchUrl argument!");

            if(string.IsNullOrEmpty(esUsername))
                throw new Exception("Missing --elasticsearchUsername argument!");

            if(string.IsNullOrEmpty(esPassword))
                throw new Exception("Missing --elasticsearchPassword argument!");


            // Begin !
            Stopwatch stopWatch = new Stopwatch();
            stopWatch.Start();
            
            // Establish connection with local Elasticsearch
            var indexMap = "variants";

            var settings = new ConnectionSettings(new Uri(url))
                .ServerCertificateValidationCallback((o, certificate, chain, errors) => true) // allow self-signed certs
                .BasicAuthentication(esUsername, esPassword)
                .DefaultIndex(indexMap);

            var client = new ElasticClient(settings);


            // Get all project vcf files
            string[] files = System.IO.Directory.GetFiles(vcfFilesPath, "*.vcf");

            int rowCount = 0;            
            ParallelOptions poFiles = new ParallelOptions {
                // Process ~3 times fewer files simultaneously 
                // as there are processors on the host machine
                MaxDegreeOfParallelism = (int)Math.Round((double)Environment.ProcessorCount / 3)
            };
            
            ParallelOptions poRows = new ParallelOptions {
                MaxDegreeOfParallelism = Environment.ProcessorCount 
            };
            
            ParallelOptions poColumns = new ParallelOptions {
                MaxDegreeOfParallelism = Environment.ProcessorCount 
            };

            Parallel.ForEach(files, poFiles, (filepath, _, fileNumber) =>
            {            
                System.Console.WriteLine("Ingesting file {0}.", filepath);

                // Collect all '#' lines as one header-block
                var headerStringBuilder = new StringBuilder();
                foreach(var line in File.ReadLines(filepath))
                {
                    if (line.ElementAt(0)=='#')
                    {
                        // Grab the Header line
                        if(line.Contains("#CHROM"))
                        {
                            break;
                        }
                        else
                        {
                            headerStringBuilder.AppendLine(line);
                        }
                    }
                }

                bool fileIndexCreateSuccess = false;
                int attempts = 0;
                IndexResponse fileResponse = new IndexResponse();
                while (fileIndexCreateSuccess == false && attempts < 100)
                {
                    System.Console.WriteLine($"Attempting to create ES index for {filepath}");

                    // Create Elasticsearch documents for the filename
                    fileResponse = client.Index(new 
                    {
                        filename = Path.GetFileName(filepath),
                        compressedHeaderBlockBase64 = Convert.ToBase64String(Utils.Zip(headerStringBuilder.ToString()))
                    }, i => i.Index("files"));

                    if (fileResponse.Id != null)
                    {
                        // Succeeded
                        fileIndexCreateSuccess = true;
                    }

                    attempts++;                    
                }

                if (fileIndexCreateSuccess == false) 
                {
                    // Abandon file
                    System.Console.WriteLine($"Failed to create ES index for file {filepath} -- Aborting this file");
                    return;
                }
                else
                {
                    System.Console.WriteLine($"Succeeded to create ES index for file {filepath} after {attempts} attempt{(attempts > 1 ? "s" : string.Empty)} : id {fileResponse.Id}");
                }


                // Create VCF documents
                List<string> headers = new List<string>();
                List<dynamic> Documents = new List<dynamic>();

                BulkDescriptor descriptor = new BulkDescriptor();
                Parallel.ForEach(File.ReadLines(filepath), poRows, (xLine, _x, lineNumber) =>
                {
                    if (xLine.ElementAt(0)=='#')
                    {
                        // Grab the Header line
                        if(xLine.Contains("CHROM"))
                        {
                            headers = xLine.Split("\t").ToList();
                        }

                        return;
                    }

                    //Dictionary<string, dynamic> doc = new Dictionary<string, dynamic>();
                    
                    // see: https://docs.microsoft.com/en-us/dotnet/api/system.collections.concurrent.concurrentdictionary-2?redirectedfrom=MSDN&view=net-5.0
                    
                    // We know how many items we want to insert into the ConcurrentDictionary.
                    // So set the initial capacity to some prime number above that, to ensure that
                    // the ConcurrentDictionary does not need to be resized while initializing it.
                    //int NUMITEMS = 64;
                    int initialCapacity = 11;

                    // The higher the concurrencyLevel, the higher the theoretical number of operations
                    // that could be performed concurrently on the ConcurrentDictionary.  However, global
                    // operations like resizing the dictionary take longer as the concurrencyLevel rises.
                    // For the purposes of this example, we'll compromise at numCores * 2.
                    int numProcs = Environment.ProcessorCount;
                    int concurrencyLevel = numProcs * 2;

                    // Construct the dictionary with the desired concurrencyLevel and initialCapacity
                    ConcurrentDictionary<string, dynamic> cd = new ConcurrentDictionary<string, dynamic>(concurrencyLevel, initialCapacity);


                    // Associate variant with its original file
                    cd["fileId"] = fileResponse.Id;

                    //List<object> docParticipantsList = new List<object>();
                    
                    // see: https://docs.microsoft.com/en-us/dotnet/api/system.collections.concurrent.concurrentbag-1?view=net-5.0
                    ConcurrentBag<object> docParticipantsBag = new ConcurrentBag<object>();

                    var rowComponents = xLine.Split("\t").ToList();// Temp cap at x //.Take(500)
                    //int columnNumber = 0;

                    List<string> defaultHeaderFields = new List<string>() 
                    {
                        "CHROM",
                        "POS",
                        "ID",
                        "REF",
                        "ALT",
                        "QUAL",
                        "FILTER",
                        "INFO",
                        "FORMAT"
                    };

                    // Dynamically generate column names and type, and add column value
                    //rowComponents.ForEach(rc =>
                    Parallel.ForEach(rowComponents, poColumns, (rc, _r, columnNumber) =>
                    {
                        try
                        {
                            // Reduce storage cost by escaping 'empty' variations
                            if (rc != "0|0")
                            {   
                                var key = headers[(int)columnNumber].Trim().Replace("#", string.Empty);
                                var value = rc.Trim();

                                if (defaultHeaderFields.Any(dhf => dhf == key))
                                {
                                    // Filter field type by column name
                                    if (string.Equals(key, "CHROM") || string.Equals(key, "POS") || string.Equals(key, "QUAL"))
                                    {
                                        int potentialIntValue;
                                        if(Int32.TryParse(value, out potentialIntValue))
                                            cd[key.ToLower()] = potentialIntValue;
                                        else
                                            cd[key.ToLower()] = -1; // equivalent to a null value (like a period '.')
                                    }
                                    // default: string
                                    else
                                        cd[key.ToLower()] = value;
                                }
                                else
                                {
                                    // Assume it's a partipant header
                                    docParticipantsBag.Add(new {
                                        SampleId = key,
                                        Variation = value
                                    });
                                }
                                
                                //columnNumber++ ;
                            }                        
                        }
                        catch(Exception ex)
                        {
                            System.Console.WriteLine($"Oops, something went wrong: {ex.Message}");
                        }
                    });

                    cd["samples"] = docParticipantsBag.ToList();

                    lock(HttpCallLockObject)
                    {
                        // Pile all documents together
                        descriptor.Index<object>(i => i
                            .Index("variants")
                            .Document(cd));

                        // Push x at a time
                        if (rowCount % 10000 == 0)
                        {
                            // TODO: check for errors
                            BulkResponse response = client.Bulk(descriptor);

                            if (response.Errors)
                            {
                                response.ItemsWithErrors.ToList().ForEach(iwe => 
                                {
                                    System.Console.WriteLine(iwe);
                                });
                            }
                            descriptor = new BulkDescriptor();

                            System.Console.Write("{0} rows ingested on so far..", rowCount);
                        }
                    }

                    rowCount++;
                });

                // Final bulk push
                BulkResponse responseX = client.Bulk(descriptor);
            });

            stopWatch.Stop();

            // Get the elapsed time as a TimeSpan value.
            TimeSpan ts = stopWatch.Elapsed;
            string elapsedTime = String.Format("{0:00}:{1:00}:{2:00}.{3:00}", 
                ts.Hours, ts.Minutes, ts.Seconds, ts.Milliseconds / 10);

            System.Console.WriteLine("Ingested {0} variant documents in time {1}.", rowCount, elapsedTime);
        }
    }
}
