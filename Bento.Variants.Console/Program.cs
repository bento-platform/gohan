﻿using System.Linq;
using System;
using System.IO;
using System.Collections.Generic;
using System.Text;
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

            // Establish connection with local Elasticsearch
            var url = "http://localhost:9200";
            var indexMap = "variants";

            var settings = new ConnectionSettings(new Uri(url))
                .DefaultIndex(indexMap);

            var client = new ElasticClient(settings);

            // Ingest 1000Genomes chr 22 into Elasticsearch

            List<string> headers = new List<string>();
            List<dynamic> Documents = new List<dynamic>();

            BulkDescriptor descriptor = new BulkDescriptor();
            int rowCount = 0;
            Parallel.ForEach(File.ReadLines("Bento.Variants.Console/ALL.chr22.phase3_shapeit2_mvncall_integrated_v5a.20130502.genotypes.vcf"), (xLine, _, lineNumber) =>
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

                Dictionary<string, dynamic> doc = new Dictionary<string, dynamic>();
                List<object> docParticipantsList = new List<object>();

                var rowComponents = xLine.Split("\t").ToList();// Temp cap at x //.Take(500)
                int columnNumber = 0;

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
                rowComponents.ForEach(rc =>
                {
                    if (rc != "0|0")
                    {   
                        var key = headers[columnNumber].Trim().Replace("#", string.Empty);
                        var value = rc.Trim(); //.Replace("|", "--");

                        if (defaultHeaderFields.Any(dhf => dhf == key))
                        {
                            // Filter field type by column name
                            if (string.Equals(key, "CHROM") || string.Equals(key, "POS") || string.Equals(key, "QUAL"))
                                doc[key] = Int32.Parse(value);
                            else
                                // default: string
                                doc[key] = value;

                        }
                        else
                        {
                            // Assume it's a partipant header
                            //doc["PARTICIPANTS"]
                            docParticipantsList.Add(new {
                                ParticipantId = key,
                                Variation = value
                            });
                        }
                        
                        columnNumber++ ;
                    }
                });

                doc["PARTICIPANTS"] = docParticipantsList.ToList();

                lock(HttpCallLockObject)
                {
                    // Pile all documents together
                    descriptor.Index<object>(i => i
                    .Index("variants")
                    .Document(doc));

                    // Push x at a time
                    if (rowCount % 10000 == 0)
                    {
                        // TODO: check for errors
                        BulkResponse response = client.Bulk(descriptor);
                        descriptor = new BulkDescriptor();

                        System.Console.WriteLine("{0} rows ingested on so far..", rowCount);
                    }
                }

                rowCount++;
            });

            // Final bulk push
            BulkResponse responseX = client.Bulk(descriptor);


            System.Console.WriteLine("Ingested {0} rows.", rowCount);
        }
    }
}
