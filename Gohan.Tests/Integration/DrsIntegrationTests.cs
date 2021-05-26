using System.Net;
using System;
using System.Threading.Tasks;
using System.Text;
using System.IO;
using System.Net.Http;

using Xunit;
using Xunit.Repeat;

using Newtonsoft.Json;

using Gohan.XCC;
using Gohan.XCC.Models.DTOs;

namespace Gohan.Tests.Integration
{
    public class DrsIntegrationTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public DrsIntegrationTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        ///<summary>
        /// Ensures that Drs is secured behind the proxy
        /// and requires a valid set of credentials by testing a 
        /// set of valid credentials against a small set of invalid credentials
        ///</summary>
        [Fact]
        public async void IsDrsAppropriatelyConfiguredWithoutBasicAuth()
        {
            bool didSucceed = false;


            try	
            {
                // Make the call
                HttpResponseMessage response = await fixture.client.GetAsync($"{fixture.PublicDrsUrl}/objects/abc123");

                var responseBody = await response.Content.ReadAsStringAsync();
                var jsonBody = Newtonsoft.Json.JsonConvert.DeserializeObject<dynamic>(responseBody);

                // Ensure response body is a json and contains a 404
                // (DRS returns this at path '/')

                if (fixture.IsApiAuthzEnabled)
                {
                    Console.WriteLine("-- API Authz is enabled, request should be blocked! --");

                    Assert.Equal(response.StatusCode, HttpStatusCode.Unauthorized);
                }
                else
                {
                    Console.WriteLine("-- API Authz is disabled, request should not be blocked, but still be a 404! --");

                    Assert.Equal(jsonBody.code.ToString(), "404");
                    Assert.Equal(jsonBody.message.ToString(), "Not Found");
                }

                
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