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

namespace Bento.Variants.Tests
{
    public class ElasticsearchTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public ElasticsearchTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        [Theory]
        [Repeat(5)]
        public async void IsElasticSearchRunningAndSecure(int x)
        {
            bool didSucceed = false;

            // Generate random texts for credentials
            string username = RandomUtil.GetRandomString();
            string password = RandomUtil.GetRandomString();

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