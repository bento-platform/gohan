using System;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

using Nest;

namespace Gohan.Api
{
    public static class ElasticSearchExtensions
    {
        public static void AddElasticSearch(
            this IServiceCollection services, IConfiguration configuration)
        {
            var host = $"{configuration["ElasticSearch:Host"]}";
            var url = $"{configuration["ElasticSearch:Protocol"]}://{host}:{configuration["ElasticSearch:Port"]}{configuration["ElasticSearch:GatewayPath"]}";
            var indexMap = configuration["ElasticSearch:PrimaryIndex"];

            var esUsername = configuration["ElasticSearch:Username"];
            var esPassword = configuration["ElasticSearch:Password"];

            if (string.IsNullOrEmpty(url) || string.IsNullOrEmpty(indexMap) || 
                string.IsNullOrEmpty(esUsername) || string.IsNullOrEmpty(esPassword) ||
                url.Contains("not-set") ||
                string.Equals(indexMap, "not-set") ||
                string.Equals(esUsername, "not-set") ||
                string.Equals(esPassword, "not-set"))
            {
                throw new Exception($"Error: Invalid Elastic Search configuration! \n" +
                    $"url: {url ?? "null"},\n " +
                    $"apiIndexMap: {indexMap ?? "null"}\n" +
                    $"-- Aborting");
            }

            Console.WriteLine("----------------");
            Console.WriteLine($"Elasticsearch URL: {url}");
# if !RELEASEALPINE
            Console.WriteLine($"Elasticsearch Credentials: {esUsername} {esPassword}");
# endif
            Console.WriteLine("----------------");

            var settings = new ConnectionSettings(new Uri(url))
                .BasicAuthentication(esUsername, esPassword)
                .DefaultIndex(indexMap)
# if DEBUG
                .EnableDebugMode()
# endif
                .ServerCertificateValidationCallback((sender, cert, chain, errors) =>
                {
                    if (cert.Subject.Contains($"CN={host}"))
                        return true;

                    Console.WriteLine($"Error - Invalid ElasticSearch SSL Certificate : Needs \"{host}\" but got \"{cert.Subject}\"");
                    return false;
                })
                ;

            var client = new ElasticClient(settings);
            services.AddSingleton<IElasticClient>(client);
        }
    }
}
