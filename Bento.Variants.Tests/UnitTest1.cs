using System.Net;
using System;
using System.Threading.Tasks;
using System.Text;
using System.IO;
using System.Net.Http;

using Xunit;
using Xunit.Repeat;

using Bento.Variants.XCC;

namespace Bento.Variants.Tests
{
    public class UnitTest1 : IClassFixture<ApiTestFixture>
    {
        public ApiTestFixture fixture;

        public UnitTest1(ApiTestFixture fixture)
        {
            this.fixture = fixture;
        }

        [Fact]
        public async void IsApiRunning()
        {
            bool didSucceed = false;

            try	
            {
                HttpResponseMessage response = await fixture.client.GetAsync(fixture.ApiUrl);

                didSucceed = response.StatusCode == HttpStatusCode.OK;
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ", e.Message);

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
                fixture.client.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Basic", Convert.ToBase64String(byteArray));

                // Make the call
                HttpResponseMessage response = await fixture.client.GetAsync($"{fixture.ApiUrl}{fixture.PublicFacingElasticPath}");

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
                Console.WriteLine("Message :{0} ", e.Message);

                didSucceed = false;
            }

            Assert.True(didSucceed);                    
        }
    }
}