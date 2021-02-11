using System.Linq;
using System;
using System.IO;
using System.Collections.Generic;
using System.Collections.Concurrent;
using System.Diagnostics;
using System.Threading.Tasks;

using Nest;

namespace Bento.Variants.Console
{
    class Program
    {
        public static object HttpCallLockObject = new object();
        static void Main(string[] args)
        {
            System.Console.WriteLine("Hello World!");

            Stopwatch stopWatch = new Stopwatch();
            stopWatch.Start();
            

            // Establish connection with local Elasticsearch
            var url = "http://localhost:9200";
            var indexMap = "variants";

            var settings = new ConnectionSettings(new Uri(url))
                .DefaultIndex(indexMap);

            var client = new ElasticClient(settings);

            // Get all project vcf files
            string[] files = System.IO.Directory.GetFiles($"{System.IO.Directory.GetCurrentDirectory()}/Bento.Variants.Console/vcf", "*.vcf");

            int rowCount = 0;
            ParallelOptions poFiles = new ParallelOptions {
                MaxDegreeOfParallelism = 2 // only x files at a time
            };
            ParallelOptions poRows = new ParallelOptions {
                MaxDegreeOfParallelism = 6
            };
            ParallelOptions poColumns = new ParallelOptions {
                MaxDegreeOfParallelism = 6
            };
            Parallel.ForEach(files, poFiles, (filepath, _, fileNumber) =>
            {            
                System.Console.WriteLine("Ingesting file {0}.", filepath);

                bool fileIndexCreateSuccess = false;
                int attempts = 0;
                IndexResponse fileResponse = new IndexResponse();
                while (fileIndexCreateSuccess == false && attempts < 100)
                {
                    System.Console.WriteLine($"Attempting to create ES index for {filepath}");

                    // Create Elasticsearch documents for the filename
                    fileResponse = client.Index(new 
                    {
                        filename = Path.GetFileName(filepath)
                    }, 
                    i => i.Index("files"));

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

                            System.Console.Write("\r{0} rows ingested on so far..", rowCount);
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
            ts.Hours, ts.Minutes, ts.Seconds,
            ts.Milliseconds / 10);

            System.Console.WriteLine("Ingested {0} variant documents in time {1}.", rowCount, elapsedTime);
        }
    }
}
