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

        public async Task<List<dynamic>> GetDocumentsContainingVariant(double chromosome, string variant, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"CHROM:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"ID:{variant}")
                            )
                        )
                    )
                )
                .Size(rowCount)
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Documents.ToList();
        }
        public async Task<List<dynamic>> GetDocumentsContainingVariantInPositionRange(double chromosome, string variant, double lowerBound, double upperBound, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"CHROM:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"ID:{variant}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("POS")
                                .GreaterThanOrEquals(lowerBound)
                                .LessThanOrEquals(upperBound)
                            )
                        )
                    )
                )
                .Size(rowCount)
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Documents.ToList();
        }

        public async Task<List<dynamic>> GetDocumentsInPositionRange(double chromosome, double lowerBound, double upperBound, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")                
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"CHROM:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("POS")
                                .GreaterThanOrEquals(lowerBound)
                                .LessThanOrEquals(upperBound)
                            )
                        )
                    )
                )
                .Size(rowCount)
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Documents.ToList();
        }


        public async Task<long> CountDocumentsContainingVariant(double chromosome, string variant)
        {
            var searchResponse = await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"CHROM:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"ID:{variant}")
                            )
                        )
                    )
                )
            );
            //var rawQuery = searchResponse.DebugInformation;
            //Console.WriteLine(rawQuery);

            return searchResponse.Count;
        }

    }
}
