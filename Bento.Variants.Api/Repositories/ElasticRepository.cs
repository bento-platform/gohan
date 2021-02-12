﻿using System.Collections.Generic;
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

        public async Task<List<dynamic>> GetDocumentsContainingSampleId(double? chromosome, string sampleId, double? lowerBound, double? upperBound, int size = 100, string sortByPosition = null)
        {
            var searchResponse = (await ElasticClient.SearchAsync<dynamic>(s => s
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
                    .Excludes(e => e
                        .Field("samples")
                    )
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

            return searchResponse.Documents.ToList();
        }


        public async Task<List<dynamic>> GetDocumentsContainingVariantId(double? chromosome, string variant, double? lowerBound, double? upperBound, int size = 100, string sortByPosition = null)
        {      
            var searchResponse = (await ElasticClient.SearchAsync<dynamic>(s => s
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

                            if (!string.IsNullOrEmpty(variant))
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"id:{variant}")
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

            return searchResponse.Documents.ToList();
        }


        public async Task<long> CountDocumentsContainingVariantInPositionRange(double? chromosome, string variant, double? lowerBound, double? upperBound)
        {      
            var searchResponse = (await ElasticClient.CountAsync<dynamic>(s => s
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

                            if (!string.IsNullOrEmpty(variant))
                            {
                                query &= fq
                                    .QueryString(d => 
                                        d.Query($"id:{variant}")
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

            return searchResponse.Count;
        }
    }
}
