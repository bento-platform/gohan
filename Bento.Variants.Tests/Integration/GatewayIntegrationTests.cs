using System.Data.Common;
using System;
using System.Net;
using System;
using System.Threading.Tasks;
using System.Text;
using System.IO;
using System.Net.Http;
using System.Linq;

using Xunit;
using Xunit.Repeat;

using Newtonsoft.Json;

using Bento.Variants.XCC;
using Bento.Variants.XCC.Models.DTOs;

namespace Bento.Variants.Tests.Integration
{
    public class GatewayIntegrationTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public GatewayIntegrationTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        ///<summary>
        /// Ensures that the services are not reachable over http and 
        /// that they force redirect to https
        ///</summary>
        [Theory]
        [InlineData("/")]
        [InlineData("/es")]
        [InlineData("/kibana")]
        public async void DoServicesOverHttpRedirectToHttps(string path)
        {
            bool didPass = false;
            try	
            {
                // Make the call
                string url = $"{fixture.InsecureVariantsGatewayUrl}{path}";
                
                using (HttpResponseMessage response = await fixture.client.GetAsync(url))
                {
                    var responseContent = response.Content.ReadAsStringAsync().Result;

                    Assert.Equal(response.StatusCode, HttpStatusCode.Moved);
                    Assert.Equal(response.Headers.Location.Scheme, "https");

                    didPass = true;
                }
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ", e.Message);
            }

            Assert.True(didPass);
        }
    }
}