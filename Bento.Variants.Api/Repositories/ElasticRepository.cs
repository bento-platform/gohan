using System;
using System.Collections.Generic;
using System.Dynamic;
using System.Linq;
using System.Net.Http;
using System.Threading.Tasks;
using System.Text;

using Microsoft.Extensions.Configuration;
using Nest;

using Bento.Variants.Api.Repositories.Interfaces;

using Bento.Variants.XCC;
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
            var updateResult = await ElasticClient.UpdateByQueryAsync<dynamic>(u => u
                .Index("variants")
                .Script(s =>s
                    .Source("ctx._source.samples.removeIf(sample -> sample.sampleId == params.sampleId)")
                    .Params(p => p
                        .Add("sampleId", sampleId)
                    ))
                .Refresh(true)
            );
            System.Console.WriteLine($"Updated : {updateResult?.OriginalException} {updateResult?.IsValid} {updateResult?.Updated}");

            var removeResult = await ElasticClient.DeleteByQueryAsync<dynamic>(u => u
                .Index("variants")
                .Query(q => q
                    .Bool(b => b
                        .MustNot(mn => mn
                            .Exists(e => e.Field("samples")))))
            );
            System.Console.WriteLine($"Deleted : {removeResult.OriginalException} {removeResult.IsValid} {removeResult?.Deleted}");

            if (updateResult?.Updated > 0 && removeResult?.Deleted == 0)
            {
                System.Console.WriteLine($"Deleted Message : {removeResult?.OriginalException?.Message}");
                System.Console.WriteLine($"Deleted InnerException Message : {removeResult?.DebugInformation}");
            }
            
        }
    }
}
