using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.IO.Compression;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading;
using System.Threading.Tasks;

using Bento.Variants.XCC;

using Nest;

using Newtonsoft.Json;

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
            
            string drsUrl = null;
            
            int documentBulkSizeLimit = 100000;

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
                        
                        case "--drsUrl":
                            if(args.Length >= argNum+1)    
                                drsUrl = $"{args[argNum+1]}";
                            break;

                        case "--bulkDocumentSizeLimit":
                            if(args.Length >= argNum+1)    
                                documentBulkSizeLimit = Int32.Parse(args[argNum+1]);
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
            

            if(string.IsNullOrEmpty(drsUrl))
                throw new Exception("Missing --drsUrl argument!");


            // documentBulkSizeLimit is optional *


            // Begin !
            Stopwatch stopWatch = new Stopwatch();
            stopWatch.Start();

            System.Console.WriteLine($"Time start : {DateTime.Now}");
            
            // Establish connection with local Elasticsearch
            var indexMap = "variants";

            var settings = new ConnectionSettings(new Uri(url))
                .ServerCertificateValidationCallback((o, certificate, chain, errors) => true) // allow self-signed certs
                .BasicAuthentication(esUsername, esPassword)
                .DefaultIndex(indexMap);

            var client = new ElasticClient(settings);


            // Create Indices
            client.Indices.Create("files");
            client.Indices.Create("variants");


            // Get all project vcf files
            string[] compressedFiles = System.IO.Directory.GetFiles(vcfFilesPath, "*.vcf.gz");

            
            // Set up parallelization configuration
            ParallelOptions poFiles, poRows,  poColumns;
                                                                    // Process ~3 times fewer files simultaneously 
                                                                    // as there are processors on the host machine
            poFiles = new ParallelOptions { MaxDegreeOfParallelism = (int)Math.Round((double)Environment.ProcessorCount / 3) };     
            poRows = poColumns = new ParallelOptions { MaxDegreeOfParallelism = Environment.ProcessorCount };

            int rowCount = 0;

            Parallel.ForEach(compressedFiles, poFiles, (compressedFilepath, _, fileNumber) =>
            {            
                System.Console.WriteLine("[{1}] Ingesting file {0}.", compressedFilepath, DateTime.Now);

                // decompress .vcf.gz files, temporarily store it to disk
                // to be processed and ingested into elasticsearch
                var decompressedFileName = Decompress(new FileInfo(compressedFilepath));

                // skip all '#' lines
                foreach(var line in File.ReadLines(decompressedFileName))
                {
                    if (line.ElementAt(0)=='#')
                    {
                        // Grab the Header line
                        if(line.Contains("#CHROM"))
                        {
                            break;
                        }
                    }
                }

                bool fileIndexCreateSuccess = false;
                int attempts = 0;
                IndexResponse fileResponse = new IndexResponse();
                string drsFileId = string.Empty;

                while (fileIndexCreateSuccess == false && attempts < 3)
                {                    
                    // TODO: ingest in DRS, and then use ID in the elasticsearch ingestion
                    HttpClientHandler httpClientHandler = new HttpClientHandler() { AllowAutoRedirect = false };
                    httpClientHandler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };

                    using (var httpClient = new HttpClient(httpClientHandler, disposeHandler: false))
                    {
                        using (var content = new MultipartFormDataContent())
                        {
                            byte[] file = System.IO.File.ReadAllBytes(compressedFilepath);
                            var filename = Path.GetFileName(compressedFilepath);

                            var byteArrayContent = new ByteArrayContent(file);

                            content.Add(byteArrayContent, "file", filename);

                            var ingestUrl = $"{drsUrl}/public/ingest";
                            var result = httpClient.PostAsync(ingestUrl, content).Result;

                            // TODO : type safety (remove dynamic, add a class)
                            var data = Newtonsoft.Json.JsonConvert.DeserializeObject<dynamic>(result.Content.ReadAsStringAsync().Result);
                            drsFileId = data.id;
                        }
                    }   

                    if (drsFileId != string.Empty )
                    {
                        // Succeeded
                        fileIndexCreateSuccess = true;
                    }
                    else
                        Thread.Sleep(3000);

                    attempts++;                    
                }

                if (fileIndexCreateSuccess == false) 
                {
                    // Abandon file
                    System.Console.WriteLine($"[{DateTime.Now}] Failed to create ES index for file {decompressedFileName} -- Aborting this file ");
                    return;
                }
                else
                {
                    System.Console.WriteLine($"[{DateTime.Now}] Succeeded to create ES index for file {decompressedFileName} after {attempts} attempt{(attempts > 1 ? "s" : string.Empty)} : id {fileResponse.Id} ");
                }


                // Create VCF documents
                List<string> headers = new List<string>();
                List<dynamic> Documents = new List<dynamic>();

                BulkDescriptor descriptor = new BulkDescriptor();
                Parallel.ForEach(File.ReadLines(decompressedFileName), poRows, (xLine, _x, lineNumber) =>
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
                    cd["fileId"] = drsFileId; // fileResponse.Id;

                    //List<object> docParticipantsList = new List<object>();
                    
                    // see: https://docs.microsoft.com/en-us/dotnet/api/system.collections.concurrent.concurrentbag-1?view=net-5.0
                    ConcurrentBag<object> docParticipantsBag = new ConcurrentBag<object>();

                    var rowComponents = xLine.Split("\t").ToList();// Temp cap at x //.Take(500)
                    //int columnNumber = 0;

                    // List<string> defaultHeaderFields = new List<string>() 
                    // {
                    //     "CHROM",
                    //     "POS",
                    //     "ID",
                    //     "REF",
                    //     "ALT",
                    //     "QUAL",
                    //     "FILTER",
                    //     "INFO",
                    //     "FORMAT"
                    // };
                    var defaultHeaderFields = Utils.VCFColumnOrder;


                    // TODO: create documents in a type-safe manner using a class structure instead of dynamically

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
                                        
                                        if (string.Equals(key, "CHROM"))
                                            value = Regex.Replace(value, "[^.0-9]", "");

                                        if(Int32.TryParse(value, out potentialIntValue))
                                            cd[key.ToLower()] = potentialIntValue;
                                        else
                                            cd[key.ToLower()] = -1; // here to simulate a null value (such as when the string value is empty, or
                                                                    // is something as arbitrary as a single period '.')
                                    }
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
                            System.Console.WriteLine($"[{DateTime.Now}] Oops, something went wrong:  \n {ex.Message}");
                        }
                    });

                    cd["samples"] = docParticipantsBag.ToList();

                    lock(HttpCallLockObject)
                    {
                        // Pile all documents together
                        descriptor.Index<object>(i => i
                            .Index("variants")
                            .Document(cd));

                        // Push "documentBulkSizeLimit" at a time
                        if (rowCount % documentBulkSizeLimit == 0)
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

                            System.Console.Write("\r[{1}] {0} rows ingested on so far..", rowCount, DateTime.Now);
                        }
                    }

                    rowCount++;
                });

                // Final bulk push
                BulkResponse responseX = client.Bulk(descriptor);


                System.IO.File.Delete(decompressedFileName);
                // TODO: load only in memory? (memory constraint concerns)
            });

            stopWatch.Stop();

            System.Console.WriteLine($"[{DateTime.Now}] DONE");

            // Get the elapsed time as a TimeSpan value.
            TimeSpan ts = stopWatch.Elapsed;
            string elapsedTime = String.Format("{0:00}:{1:00}:{2:00}.{3:00}", 
                ts.Hours, ts.Minutes, ts.Seconds, ts.Milliseconds / 10);

            System.Console.WriteLine("Ingested {0} variant documents in time {1}.", rowCount, elapsedTime);
        }
    
        public static string Decompress(FileInfo fileToDecompress)
        {
            using (FileStream originalFileStream = fileToDecompress.OpenRead())
            {
                string currentFileName = fileToDecompress.FullName;
                string newFileName = currentFileName.Remove(currentFileName.Length - fileToDecompress.Extension.Length);

                using (FileStream decompressedFileStream = File.Create(newFileName))
                {
                    using (GZipStream decompressionStream = new GZipStream(originalFileStream, CompressionMode.Decompress))
                    {
                        decompressionStream.CopyTo(decompressedFileStream);
                        System.Console.WriteLine($"Decompressed: {fileToDecompress.Name}");

                        return newFileName;
                    }
                }
            }
        }
    }


}
