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
    public class ApiTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public ApiTests(IntegrationTestFixture fixture)
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

        [Fact]
        public async void CanGetVariantsBaseLineQuery()
        {
            bool didSucceed = false;

            try	
            {
                // Make the call
                var url = $"{fixture.ApiUrl}{fixture.GetVariantsByVariantIdPath}";
                
                using (HttpResponseMessage response = await fixture.client.GetAsync(url))
                {
                    var responseContent = response.Content.ReadAsStringAsync().Result;
                    response.EnsureSuccessStatusCode();

                    var data = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);

                    Assert.Equal(response.StatusCode, HttpStatusCode.OK);

                    Assert.Equal(data.Status, 200);
                    Assert.Equal(data.Message, "Success");

                    didSucceed = true;
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