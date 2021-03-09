using System;
using System.Net.Http;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Text;

using Microsoft.Extensions.Configuration;
using Nest;

using Bento.Variants.Api.Repositories.Interfaces;

using Bento.Variants.XCC.Models.Indexes;

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

        public async Task<List<VariantIndex>> GetDocumentsContainingVariantOrSampleIdInPositionRange(long? chromosome, 
            long? lowerBound, long? upperBound, 
            string variantId = null, string sampleId = null, 
            int size = 100, string sortByPosition = null,
            bool includeSamplesInResultSet = true)
        {
            var searchResponse = (await ElasticClient.SearchAsync<VariantIndex>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(bq => bq
                        .Filter(fq =>
                        {
                            QueryContainer query = null;

                            if (chromosome.HasValue && chromosome.Value > 0)
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"chrom:{chromosome}")
                                    );
                            }

                            if (!string.IsNullOrEmpty(variantId))
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"id:{variantId}")
                                    );
                            }

                            if (!string.IsNullOrEmpty(sampleId))
                            {
                                query &= fq
                                    .Match(m => m
                                        .Field("samples.sampleId")
                                        .Query($"{sampleId}")
                                    );
                            }

                            if (lowerBound.HasValue)
                            {
                                query &= fq
                                    .Range(r => r
                                        .Field("pos")
                                        .GreaterThanOrEquals(lowerBound)
                                );
                            }

                            if (upperBound.HasValue)
                            {
                                query &= fq
                                    .Range(r => r
                                        .Field("pos")
                                        .LessThanOrEquals(upperBound)
                                );
                            }

                            return query;
                        }
                    )                    
                ))
                .Source(src => src
                    .IncludeAll()
                    .Excludes(e => 
                    {
                        if (includeSamplesInResultSet == false)
                        {
                            e.Field("samples");
                        }
                        return e;
                    })
                )
                .Sort(sort => sort
                    .Field(f =>
                    {
                        f.Field("pos");

                        if (!string.IsNullOrEmpty(sortByPosition))
                        {

                            switch (sortByPosition)
                            {
                                case "asc":
                                    f.Order(Nest.SortOrder.Ascending);
                                    break;
                                case "desc":
                                    f.Order(Nest.SortOrder.Descending);
                                    break;
                                default:
                                    break;
                            }
                        }

                        return f;           
                    }))
                .Size(size))
            );
            
            //var rawQuery = searchResponse.DebugInformation;
            //System.Console.WriteLine(rawQuery);

            if (!searchResponse.IsValid)
                throw new System.Exception("Cannot connect to Elasticsearch!");
            
            return searchResponse.Documents.ToList();
        }

        public async Task<long> CountDocumentsContainingVariantOrSampleIdInPositionRange(long? chromosome,
            long? lowerBound, long? upperBound,
            string variantId = null, string sampleId = null)
        {      
            var countResponse = (await ElasticClient.CountAsync<dynamic>(s => s
                .Index($"{Configuration["PrimaryIndex"]}")
                .Query(q => q
                    .Bool(bq => bq
                        .Filter(fq =>
                        {
                            QueryContainer query = null;

                            if (chromosome.HasValue && chromosome.Value > 0)
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"chrom:{chromosome}")
                                    );
                            }

                            if (!string.IsNullOrEmpty(variantId))
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"id:{variantId}")
                                    );
                            }

                            if (!string.IsNullOrEmpty(sampleId))
                            {
                                query &= fq
                                    .Match(m => m
                                        .Field("samples.sampleId")
                                        .Query($"{sampleId}")
                                    );
                            }

                            if (lowerBound.HasValue)
                            {
                                query &= fq
                                    .Range(r => r
                                        .Field("pos")
                                        .GreaterThanOrEquals(lowerBound)
                                );
                            }

                            if (upperBound.HasValue)
                            {
                                query &= fq
                                    .Range(r => r
                                        .Field("pos")
                                        .LessThanOrEquals(upperBound)
                                );
                            }

                            return query;
                        }
                    )                    
                )))
            );

            //var rawQuery = searchResponse.DebugInformation;
            //System.Console.WriteLine(rawQuery);

            if (!countResponse.IsValid)
                throw new System.Exception("Cannot connect to Elasticsearch!");
            
            return countResponse.Count;
        }
    
        public async Task<dynamic> GetFileByFileId(string fileId)
        {
            var searchResponse = (await ElasticClient.SearchAsync<dynamic>(s => s
                .Index($"files")
                .Query(q => q
                    .Bool(bq => bq
                        .Filter(fq =>
                        {
                            QueryContainer query = null;

                            if (!string.IsNullOrEmpty(fileId))
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"_id:{fileId}")
                                    );
                            }
                            return query;
                        }
                    )                    
                ))
                .Source(src => src
                    .IncludeAll()
                ))
            );
            
            //var rawQuery = searchResponse.DebugInformation;
            //System.Console.WriteLine(rawQuery);

            if (!searchResponse.IsValid)
                throw new System.Exception("Cannot connect to Elasticsearch!");
            
            return searchResponse.Documents.FirstOrDefault();
        }
    
        public async Task RemoveSampleFromVariantsBySampleId(string sampleId)
        {
            var host = $"{Configuration["ElasticSearch:Host"]}";
            var indexMap = Configuration["ElasticSearch:PrimaryIndex"];

            var esUsername = Configuration["ElasticSearch:Username"];
            var esPassword = Configuration["ElasticSearch:Password"];
            
            var baseUrl = $"{Configuration["ElasticSearch:Protocol"]}://{host}:{Configuration["ElasticSearch:Port"]}{Configuration["ElasticSearch:GatewayPath"]}";

            // Update
            // TODO: optimize (very slow)
            var updatePath = "/variants/_update_by_query?conflicts=proceed";
            var updatePayload = $@"{{
                ""script"":{{
                    ""source"": ""ctx._source.samples.removeIf(sample -> sample.sampleId == params.sampleId)"",
                    ""params"": {{
                        ""sampleId"": ""{sampleId}""
                    }}
                }}
            }}";
            var updateUrl = $"{baseUrl}{updatePath}";

            // Delete variant document if samples are empty
            var deletePath = "/variants/_delete_by_query";
            var deletePayload = @"
            {
                ""query"": {
                    ""bool"": {
                        ""should"": [
                            {
                                ""bool"": {
                                    ""must_not"": [
                                        {
                                            ""exists"": {
                                                ""field"": ""samples""
                                            }
                                        }
                                    ]
                                }
                            }
                        ]
                    }
                }
            }";
            var deleteUrl = $"{baseUrl}{deletePath}";

            using (HttpClientHandler handler = new HttpClientHandler())
            {
#if DEBUG
            handler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };
#endif
                using (HttpClient client = new HttpClient(handler, disposeHandler: false))
                {
                    // Basic Auth
                    var byteArray = Encoding.ASCII.GetBytes($"{esUsername}:{esPassword}");
                    client.DefaultRequestHeaders.Authorization = 
                        new System.Net.Http.Headers.AuthenticationHeaderValue("Basic", Convert.ToBase64String(byteArray));

                    // Remove Samples from Variants 
                    using (HttpResponseMessage response = await client.PostAsync(updateUrl, new StringContent(updatePayload, Encoding.UTF8, "application/json")))
                    {
                        var responseContent = response.Content.ReadAsStringAsync().Result;
                        System.Console.WriteLine(responseContent);
                    }

                    // Remove Variants that have no samples
                    using (HttpResponseMessage response = await client.PostAsync(deleteUrl, new StringContent(deletePayload, Encoding.UTF8, "application/json")))
                    {
                        var responseContent = response.Content.ReadAsStringAsync().Result;
                        System.Console.WriteLine(responseContent);
                    }
                }
            }
        }
    }
}
