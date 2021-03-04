using System.Net;
using System;
using System.Net.Http;
using System.Threading.Tasks;
using System.Text;
using System.IO;

using Xunit;
using Xunit.Repeat;

using Bento.Variants.XCC;

namespace Bento.Variants.Tests
{
    public class UnitTest1
    {
        // TODO: Get from .env
        public string ApiUrl = "https://variants.local:443";

        public HttpClient client;
        public HttpClientHandler httpClientHandler = new HttpClientHandler();


        public UnitTest1()
        {
            // Set up UnitTest1-wide http client
#if DEBUG
            httpClientHandler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };
            client = new HttpClient(httpClientHandler);
# else
            client = new HttpClient();
#endif

            client.Timeout = TimeSpan.FromSeconds(3);
        }

        [Fact]
        public async void IsApiRunning()
        {
            bool didSucceed = false;

            try	
            {
                HttpResponseMessage response = await client.GetAsync(ApiUrl);

                didSucceed = response.StatusCode == HttpStatusCode.OK;
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ",e.Message);

                didSucceed = false;
            }

            Assert.True(didSucceed);                    
        }

        [Theory]
        [Repeat(10)]
        public async void IsElasticSearchRunningAndSecure(int x)
        {
            bool didSucceed = false;

            // Generate random texts for credentials
            string username = RandomUtil.GetRandomString();
            string password = RandomUtil.GetRandomString();

            if (x == 1)
            {
                // TODO: get from .env
                // ** doens't work yet - will get 1 failed test
                username = Environment.GetEnvironmentVariable("DOTNET_TEST_USERNAME");
                password = Environment.GetEnvironmentVariable("DOTNET_TEST_PASSWORD");

                var m = $"Testing actual Elasticsearch credentials..";
#if DEBUG
                m += $" {username} {password}";
#endif
                Console.WriteLine(m);
            }

            try	
            {
                // Create Basic Authentication header
                var byteArray = Encoding.ASCII.GetBytes($"{username}:{password}");
                client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Basic", Convert.ToBase64String(byteArray));

                // Make the call
                HttpResponseMessage response = await client.GetAsync($"{ApiUrl}/es");

                // Ensure random credentials are blocked
                didSucceed = response.StatusCode == HttpStatusCode.Unauthorized;

                if (x == 1)
                {
                    // .. and that the actual credentials get through
                    didSucceed = response.StatusCode == HttpStatusCode.OK;
                }
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ",e.Message);

                didSucceed = false;
            }

            Assert.True(didSucceed);                    
        }
    }
}