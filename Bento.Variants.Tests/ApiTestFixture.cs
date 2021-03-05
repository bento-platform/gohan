using System;
using System.Net.Http;

using Microsoft.Extensions.Configuration;

namespace Bento.Variants.Tests
{
    public class ApiTestFixture : IDisposable
    {
        public string ApiUrl;
        public string PublicFacingElasticPath;

        public string ElasticUsername;
        public string ElasticPassword;

        public HttpClient client;
        public HttpClientHandler httpClientHandler = new HttpClientHandler();


        public ApiTestFixture()
        {
            // Load Configuration
            var config = new ConfigurationBuilder()
                    .AddJsonFile("appsettings.test.json")
                    .Build();

            ApiUrl = config["ApiUrl"];
            PublicFacingElasticPath = config["PublicFacingElasticPath"];

            ElasticUsername = config["ElasticUsername"];
            ElasticPassword = config["ElasticPassword"];

            
            // Set up UnitTest1-wide http client
#if DEBUG
            httpClientHandler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };
            client = new HttpClient(httpClientHandler);
# else
            client = new HttpClient();
#endif

            client.Timeout = TimeSpan.FromSeconds(3);

        }

        public void Dispose()
        {
            // ... clean up test data from the database ...
        }
    }
}