using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using Microsoft.Extensions.Configuration;
using Nest;

using Bento.Variants.Api.Repositories.Interfaces;

namespace Bento.Variants.Api.Repositories
{
    public class ElasticRepository : IElasticRepository
    {
        private readonly IConfiguration Configuration;
        private readonly IElasticClient ElasticClient;

        public ElasticRepository(IConfiguration configuration, IElasticClient elasticClient)
        {
            Configuration = configuration;
            ElasticClient = elasticClient;
        }

        // public void AddClientCsvFromUpload(Guid uploadId, byte[] filebytes)
        // {
        //     List<dynamic> Documents = new List<dynamic>();
        //     //Populate Documents

        //     BulkDescriptor descriptor = new BulkDescriptor();

        //     // Dynamically generate column names and type, and add column value
        //     // ...

        //     // Load file
        //     var csvStr = Encoding.UTF8.GetString(filebytes);


        //     using (var reader = new StringReader(csvStr))
        //     using (var csvReader = new CsvReader(reader, System.Globalization.CultureInfo.CreateSpecificCulture("enUS")))
        //     {
        //         List<string> badRecord = new List<string>();
        //         csvReader.Configuration.BadDataFound = context => badRecord.Add(context.RawRecord);
        //         csvReader.Configuration.PrepareHeaderForMatch = (header, i) =>
        //         {
        //             return Regex.Replace(header, @"\s", string.Empty);
        //         };

        //         // Verify Column Names
        //         csvReader.Read();
        //         csvReader.ReadHeader();
        //         List<string> headerRowColumnNames = csvReader.Context.HeaderRecord
        //             .Select(hr => hr.Trim())
        //             .ToList();

        //         // Get Records
        //         while (csvReader.Read())
        //         {
        //             Dictionary<string, dynamic> doc = new Dictionary<string, dynamic>();

        //             int columnNumber = 1;
        //             headerRowColumnNames.ForEach(hr =>
        //             {
        //                 doc[$"column_{columnNumber}_name"] = hr;

        //                 dynamic otherKindOfValue;
        //                 double? numericValue;

        //                 if (csvReader.TryGetField<double?>(hr, out numericValue))
        //                 {
        //                     doc[$"column_{columnNumber}_type"] = "numeric";
        //                     doc[$"column_{columnNumber}_value"] = numericValue;
        //                 }
        //                 else if (csvReader.TryGetField<dynamic>(hr, out otherKindOfValue))
        //                 {
        //                     doc[$"column_{columnNumber}_type"] = "other";
        //                     doc[$"column_{columnNumber}_value"] = otherKindOfValue;
        //                 }

        //                 columnNumber++ ;
        //             });

        //             descriptor.Index<object>(i => i
        //             .Index($"{Configuration["ElasticSearch:CsvIndex"]}-{uploadId}")
        //             .Document(doc));
        //         }
        //     }

        //     ElasticClient.Bulk(descriptor);            
        // }

        // public List<dynamic> GetCsvRowsByUploadId(Guid uploadId, int rowCount)
        // {
        //     var searchResponse = ElasticClient.Search<dynamic>(s => s
        //         .Index($"{Configuration["ElasticSearch:CsvIndex"]}-{uploadId}")
        //         .Query(q => q
        //             .MatchAll()
        //         )
        //         .Size(rowCount)
        //     );

        //     return searchResponse.Documents.ToList();
        // }
        public bool SimulateElasticSearchGet()
        {
            return true;
        }

        public void SimulateElasticSearchSet(int x)
        {
            Console.WriteLine($"Setting {x}");
        }
    }
}
