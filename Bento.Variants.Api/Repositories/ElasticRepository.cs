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

        public async Task<List<dynamic>> GetDocumentsContainingVariant(string chromosome, string variant, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"id:{variant}")
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

        public async Task<List<dynamic>> GetDocumentsContainingVariantInPositionRange(string chromosome, string variant, double lowerBound, double upperBound, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"id:{variant}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("pos")
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

        public async Task<List<dynamic>> GetDocumentsInPositionRange(string chromosome, double lowerBound, double upperBound, int rowCount = 100)
        {
            var searchResponse = await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")                
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("pos")
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


        public async Task<long> CountDocumentsContainingVariant(string chromosome, string variant)
        {
            var searchResponse = await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"id:{variant}")
                            )
                        )
                    )
                )
            );

            return searchResponse.Count;
        }

        public async Task<long> CountDocumentsContainingVariantInPositionRange(string chromosome, string variant, double lowerBound, double upperBound)
        {
            var searchResponse = await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"id:{variant}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("pos")
                                .GreaterThanOrEquals(lowerBound)
                                .LessThanOrEquals(upperBound)
                            )
                        )
                    )
                )
            );

            return searchResponse.Count;
        }

        public async Task<long> CountDocumentsInPositionRange(string chromosome, double lowerBound, double upperBound)
        {
            var searchResponse = await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")                
                .Query(q => q
                    .Bool(b => b
                        .Must(m => m
                            .QueryString(d => 
                                d.Query($"chrom:{chromosome}")
                            )
                        )
                        .Must(m => m
                            .Range(r => r
                                .Field("pos")
                                .GreaterThanOrEquals(lowerBound)
                                .LessThanOrEquals(upperBound)
                            )
                        )
                    )
                )
            );

            return searchResponse.Count;
        }
    }
}
