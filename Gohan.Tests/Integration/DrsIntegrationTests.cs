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
        [Theory]
        [Repeat(10)]
        public async void IsDrsRunningAndSecure(int x)
        {
            bool didSucceed = false;

            // Generate random texts for credentials
            string username = RandomUtil.GetRandomString();
            string password = RandomUtil.GetRandomString();

            // First test the valid credentials
            if (x == 1)
            {
                username = fixture.DrsUsername;
                password = fixture.DrsPassword;

                var m = $"Testing actual Drs credentials..";
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
                HttpResponseMessage response = await fixture.client.GetAsync($"{fixture.VariantsGatewayUrl}{fixture.PublicFacingDrsPath}");

                if (x == 1)
                {
                    var responseBody = await response.Content.ReadAsStringAsync();
                    var jsonBody = Newtonsoft.Json.JsonConvert.DeserializeObject<dynamic>(responseBody);

                    // Ensure response body is a json and contains a 404
                    // (DRS returns this at path '/')
                    Assert.Equal(jsonBody.code.ToString(), "404");
                    Assert.Equal(jsonBody.message.ToString(), "Not Found");
                }
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