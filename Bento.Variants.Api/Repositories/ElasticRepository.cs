using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
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

        public async Task<List<dynamic>> GetDocumentsContainingVariant(string variant, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .QueryString(d => 
                        d.Query($"{variant}:*")))
                .Size(rowCount)
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Documents.ToList();
        }

        public async Task<long> CountDocumentsContainingVariant(string variant)
        {
            var searchResponse = await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .QueryString(d => 
                        d.Query($"{variant}:*")))
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Count;
        }

    }
}
