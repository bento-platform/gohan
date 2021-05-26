using System;
using System.Net.Http;

using Microsoft.Extensions.Configuration;

namespace Gohan.Tests
{
    public class IntegrationTestFixture : IDisposable
    {
        public bool IsApiAuthzEnabled;
        
        public string VariantsGatewayUrl;
        public string InsecureVariantsGatewayUrl;

        public string GetVariantsByVariantIdPath = "/variants/get/by/variantId";
        public string GetVariantsBySampleIdPath = "/variants/get/by/sampleId";
        public string CountVariantsByVariantIdPath = "/variants/count/by/variantId";
        public string CountVariantsBySampleIdPath = "/variants/count/by/sampleId";

        public string RemoveSampleIdPath = "/variants/remove/sampleId";


        public string PublicElasticUrl;

        public string ElasticUsername;
        public string ElasticPassword;


        public string PublicDrsUrl;
        
        public string DrsUsername;
        public string DrsPassword;


        public HttpClient client;
        public HttpClientHandler httpClientHandler = new HttpClientHandler() { AllowAutoRedirect = false };

        public IntegrationTestFixture()
        {
            // Load Configuration
            var config = new ConfigurationBuilder()
                .AddJsonFile("appsettings.test.json")
                .Build();

            // Set up test-wide http configuration
            IsApiAuthzEnabled = Boolean.Parse(config["IsApiAuthzEnabled"]);

            VariantsGatewayUrl = config["VariantsGatewayUrl"];
            InsecureVariantsGatewayUrl = config["InsecureVariantsGatewayUrl"];


            PublicElasticUrl = config["PublicElasticUrl"];

            ElasticUsername = config["ElasticUsername"];
            ElasticPassword = config["ElasticPassword"];


            PublicDrsUrl = config["PublicDrsUrl"];

            DrsUsername = config["DrsUsername"];
            DrsPassword = config["DrsPassword"];


#if DEBUG
            httpClientHandler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };
#endif
            client = new HttpClient(httpClientHandler, disposeHandler: false);
            client.Timeout = TimeSpan.FromMinutes(10);
        }

        public void Dispose()
        {
            // ... clean up test data from the database ...
        }
    }
}