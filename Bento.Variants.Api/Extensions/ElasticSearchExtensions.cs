using System;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Nest;

namespace Bento.Variants.Api
{
    public static class ElasticSearchExtensions
    {
        public static void AddElasticSearch(
            this IServiceCollection services, IConfiguration configuration)
        {
            var url = $"{configuration["ElasticSearch:Protocol"]}://{configuration["ElasticSearch:Host"]}:{configuration["ElasticSearch:Port"]}";
            var indexMap = configuration["ElasticSearch:PrimaryIndex"];

            if (string.IsNullOrEmpty(url) || string.IsNullOrEmpty(indexMap) ||
                url.Contains("not-set") ||
                string.Equals(indexMap, "not-set"))
            {
                throw new Exception($"Error: Invalid Elastic Search configuration! \n" +
                    $"url: {url ?? "null"},\n " +
                    $"apiIndexMap: {indexMap ?? "null"}\n" +
                    $"-- Aborting");
            }

            var settings = new ConnectionSettings(new Uri(url))
                .BasicAuthentication($"{configuration["ElasticSearch:Username"]}", $"{configuration["ElasticSearch:Password"]}")
                .DefaultIndex(indexMap)
                //.EnableDebugMode()
                //.DefaultMappingFor<dynamic>(m => m
                //    .PropertyName(p => p.UnixTimestampUTC, "unixtimestamputc")
                //    .PropertyName(p => p.Message, "message"))
                // .ServerCertificateValidationCallback((sender, cert, chain, errors) =>
                // {
                //     if (cert.Subject == "CN=variants.local")
                //         return true;

                //     Console.WriteLine($"Error - Invalid ElasticSearch SSL Certificate : Subject {cert.Subject}");
                //     return false;
                // })
                ;

            var client = new ElasticClient(settings);

            // // Map and Ensure Minote Load-Balancer Logs Index is created
            // var createIndexResponse = client.Indices.Create("minote-load-balancer-logs", c => c
            //     .Map<MinoteLoadBalancerIndexMap>(m => m)
            // );

            services.AddSingleton<IElasticClient>(client);
        }
    }
}
