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
        /// Ensures that variants can be queried for by
        /// pinging `/variants/get/by/variantId` without any parameters
        /// and verifying the existance of results
        ///</summary>
        [Fact]
        public async void CanGetVariantsBaseLineQuery()
        {
            bool didSucceed = false;

            try	
            {
                // Make the call
                string url = $"{fixture.ApiUrl}{fixture.GetVariantsByVariantIdPath}";
                
                using (HttpResponseMessage response = await fixture.client.GetAsync(url))
                {
                    var responseContent = response.Content.ReadAsStringAsync().Result;

                    Assert.Equal(response.StatusCode, HttpStatusCode.OK);
            
                    var dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);

                    Assert.True(dto != null);

                    Assert.Equal(dto.Status, 200);
                    Assert.Equal(dto.Message, "Success");
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

        ///<summary>
        /// Ensures the order in which variants data is returned is in 
        /// ascending order by variant position
        ///</summary>
        [Fact]
        public async void CanGetVariantsInAscendingOrder()
        {
            var dto = await GetVariantsInOrder("asc");

            Assert.True(dto != null);

            Assert.Equal(dto.Status, 200);
            Assert.Equal(dto.Message, "Success");

            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null, "Data is null!");
            
            var expectedList = data.Results.OrderBy(x => x["pos"]);

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

            Assert.True(dto != null);

            Assert.Equal(dto.Status, 200);
            Assert.Equal(dto.Message, "Success");

            var data = dto.Data.FirstOrDefault();

            Assert.True(data != null);
            
            var expectedList = data.Results.OrderByDescending(x => x["pos"]);

            Assert.True(
                expectedList.SequenceEqual(data.Results), 
                "The list is not in descending order!");
        }

        ///<summary>
        /// Common function to handle ordered-by-position variants calls 
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
    }
}