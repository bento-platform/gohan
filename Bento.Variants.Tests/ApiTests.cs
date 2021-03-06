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

namespace Bento.Variants.Tests
{
    public class ApiTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public ApiTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        ///<summary>
        /// Ensures that the api is reachable by pinging / and checking for a 200 status code
        ///</summary>
        [Fact]
        public async void IsApiRunning()
        {
            bool didSucceed = false;

            try	
            {
                HttpResponseMessage response = await fixture.client.GetAsync(fixture.ApiUrl);

                Assert.Equal(response.StatusCode, HttpStatusCode.OK);

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

        
        ///<summary>
        /// Ensures that variants can be queried for and to include samples in the results set by
        /// pinging `/variants/get/by/variantId and verifying results
        ///</summary>
        [Fact]
        public async void CanGetVariantsWithSamplesInResultset()
        {
            var dto = await GetVariantsWithSamplesInResultset(true);            
            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null, "Data is null!");
            
            var expectedList = data.Results.First().Samples;

            // Ensure samples were actually returned
            Assert.True(expectedList != null, "Samples are null!");
            Assert.True(expectedList.Count > 0, "Samples are missing!");

            // Validate the state of the variations
            Assert.True(
                expectedList.All(x => 
                    x != null &&
                    x.Variation != null &&
                    x.Variation != string.Empty),
                "Variations are empty");
        }

        ///<summary>
        /// Ensures that variants can be queried for while excluding samples from the result set
        ///</summary>
        [Fact]
        public async void CanGetVariantsWithOutSamplesInResultset()
        {
            var dto = await GetVariantsWithSamplesInResultset(false);            
            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null, "Data is null!");
            
            // Make sure there aren't any samples present
            var expectedNullSamples = data.Results.First().Samples;

            Assert.True(expectedNullSamples == null, "Samples are not null!");
        }


        
    


        ///<summary>
        /// Ensures the order in which variants data is returned is in 
        /// ascending order by variant position
        ///</summary>
        [Fact]
        public async void CanGetVariantsInAscendingOrder()
        {
            var dto = await GetVariantsInOrder("asc");

            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null, "Data is null!");
            
            var expectedList = data.Results.OrderBy(x => x.Pos);

            Assert.True(
                expectedList.SequenceEqual(data.Results), 
                "The list is not in ascending order!");
        }


        ///<summary>
        /// Ensures the order in which variants data is returned is in 
        /// descending order by variant position
        ///</summary>
        [Fact]
        public async void CanGetVariantsInDesendingOrder()
        {
            var dto = await GetVariantsInOrder("desc");

            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null);
            
            var expectedList = data.Results.OrderByDescending(x => x.Pos);

            Assert.True(
                expectedList.SequenceEqual(data.Results), 
                "The list is not in descending order!");
        }


        ///<summary>
        /// Common function to handle ordered-by-position get-variants calls 
        ///</summary>
        private async Task<VariantsResponseDTO> GetVariantsInOrder(string order)
        {
            try	
            {
                VariantsResponseDTO dto;
                
                // Make the call
                string query = $"?sortByPosition={order}";
                string url = $"{fixture.ApiUrl}{fixture.GetVariantsByVariantIdPath}{query}";
                
                using (HttpResponseMessage response = await fixture.client.GetAsync(url))
                {
                    var responseContent = response.Content.ReadAsStringAsync().Result;

                    Assert.Equal(response.StatusCode, HttpStatusCode.OK);

                    dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);
                    
                    Assert.True(dto != null);

                    Assert.Equal(dto.Status, 200);
                    Assert.Equal(dto.Message, "Success");
                }
                
                return dto;
            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ", e.Message);

                return null;
            }
        }

        ///<summary>
        /// Common function to handle including/excluding samples from variants calls
        ///</summary>
        private async Task<VariantsResponseDTO> GetVariantsWithSamplesInResultset(bool includeSamples)
        {
            VariantsResponseDTO dto = null;

            try	
            {
                // Make the call
                string query = $"?includeSamplesInResultSet={includeSamples}";
                string url = $"{fixture.ApiUrl}{fixture.GetVariantsByVariantIdPath}{query}";
                
                using (HttpResponseMessage response = await fixture.client.GetAsync(url))
                {
                    var responseContent = response.Content.ReadAsStringAsync().Result;

                    Assert.Equal(response.StatusCode, HttpStatusCode.OK);
            
                    dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);

                    Assert.True(dto != null);

                    Assert.Equal(dto.Status, 200);
                    Assert.Equal(dto.Message, "Success");
                }

            }
            catch(HttpRequestException e)
            {
                Console.WriteLine("\nException Caught!");	
                Console.WriteLine("Message :{0} ", e.Message);
            }

            return dto;                   
        }
    }
}