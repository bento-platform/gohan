using System.Net;
using System;
using System.Threading.Tasks;
using System.Text;
using System.IO;
using System.Net.Http;

using Xunit;
using Xunit.Repeat;

using Newtonsoft.Json;

using Bento.Variants.XCC;
using Bento.Variants.XCC.Models.DTOs;

namespace Bento.Variants.Tests.Integration
{
    public class ElasticsearchIntegrationTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public ElasticsearchIntegrationTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        ///<summary>
        /// Ensures that Elasticsearch is secured behind the proxy
        /// and requires a valid set of credentials by testing a 
        /// set of valid credentials against a small set of invalid credentials
        ///</summary>
        [Theory]
        [Repeat(10)]
        public async void IsElasticSearchRunningAndSecure(int x)
        {
            bool didSucceed = false;

            // Generate random texts for credentials
            string username = RandomUtil.GetRandomString();
            string password = RandomUtil.GetRandomString();

            // First test the valid credentials
            if (x == 1)
            {
                username = fixture.ElasticUsername;
                password = fixture.ElasticPassword;

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
                
                fixture.client.DefaultRequestHeaders.Authorization = 
                    new System.Net.Http.Headers.AuthenticationHeaderValue("Basic", Convert.ToBase64String(byteArray));

                // Make the call
                HttpResponseMessage response = await fixture.client.GetAsync($"{fixture.ApiUrl}{fixture.PublicFacingElasticPath}");

                if (x == 1)    
                    // Ensure actual credentials get through
                    Assert.Equal(response.StatusCode, HttpStatusCode.OK);
                
                else
                    // Ensure random credentials are blocked
                    Assert.Equal(response.StatusCode, HttpStatusCode.Unauthorized);

                didSucceed = true;
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ", e.Message);

                didSucceed = false;
            }

            Assert.True(didSucceed);                    
        }
    }
}