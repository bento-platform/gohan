using System;
using System.Collections.Generic;
using System.Data.Common;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

using Xunit;
using Xunit.Repeat;

using Nest;
using Newtonsoft.Json;

using Gohan.XCC;
using Gohan.XCC.Models;
using Gohan.XCC.Models.DTOs;
using Gohan.XCC.Models.Indexes;

namespace Gohan.Tests.Integration
{
    public class ApiIntegrationTests : IClassFixture<IntegrationTestFixture>
    {
        public IntegrationTestFixture fixture;

        public ApiIntegrationTests(IntegrationTestFixture fixture)
        {
            this.fixture = fixture;
        }

        ///<summary>
        /// Ensures that the api is reachable by pinging `/` and checking for a 200 status code
        ///</summary>
        [Fact]
        public async void IsApiRunning()
        {
            bool didSucceed = false;

            try	
            {
                HttpResponseMessage response = await fixture.client.GetAsync(fixture.VariantsGatewayUrl);

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
            
            var expectedList = data.Results.FirstOrDefault()?.Samples;

            // Ensure samples were actually returned
            Assert.True(expectedList != null, "Samples are null!");
            Assert.True(expectedList.Count > 0, "Samples are missing!");
            
            // Check the sample data
            Assert.True(dto.Data.All(x => 
                x.Results.All(y => 
                    y != null &&
                    y.Samples.All(z =>
                        z != null && 
                        !string.IsNullOrEmpty(z.SampleId) &&
                        !string.IsNullOrEmpty(z.Variation)
                ))), "Something's wrong with the samples! samples are null, or variations are null/empty!");
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
            var expectedNullSamples = data.Results.FirstOrDefault()?.Samples;

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

//         [Fact]
//         public void CanAddVariantWithSample()
//         {        
//             var username = fixture.ElasticUsername;
//             var password = fixture.ElasticPassword;

//             var url = $"{fixture.VariantsGatewayUrl}{fixture.PublicFacingElasticPath}";

//             // Create ES Client with Basic Authentication header
//             var settings = new ConnectionSettings(new Uri(url))
//                 .BasicAuthentication(username, password)
//                 .DefaultIndex("variants")
// # if DEBUG
//                 .EnableDebugMode()
// # endif
//                 .ServerCertificateValidationCallback((sender, cert, chain, errors) => { return true; });

//             var client = new ElasticClient(settings);

//             var indexResponse = AddVariantWithSample(client);


//             // Check results
//             Assert.Equal(indexResponse.Id, "test-id");
//             Assert.True(indexResponse.IsValid);

//             // Verify by retrieving test doc
//             string searchQuery = $"?ids=test-id";
//             string searchUrl = $"{fixture.VariantsGatewayUrl}{fixture.GetVariantsByVariantIdPath}{searchQuery}";
            
//             using (HttpResponseMessage response = fixture.client.GetAsync(searchUrl).Result)
//             {
//                 var responseContent = response.Content.ReadAsStringAsync().Result;

//                 Assert.Equal(response.StatusCode, HttpStatusCode.OK);

//                 var dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);
                
//                 Assert.True(dto != null);

//                 Assert.Equal(dto.Status, 200);
//                 Assert.Equal(dto.Message, "Success");

//                 System.Console.WriteLine(Newtonsoft.Json.JsonConvert.SerializeObject(dto));

//                 Assert.True(dto.Data
//                     .First(d => d.VariantId == "test-id")
//                     .Results.Count == 1);
//             }
//         }

        private IndexResponse AddVariantWithSample(ElasticClient client)
        {
            // Insert dummy variant
            return client.Index(new VariantIndex()
            {
                Chrom = 0,
                Pos = 0,
                Id = "test-id",
                Ref = "test-ref",
                Alt = "test-alt",
                Qual = 0,
                Filter = "test-filter",
                Info = "test-info",
                Format = "test-format",
                FileId = "test-fileId",

                Samples = new List<Sample> 
                {
                    new Sample() { SampleId = "test-sampleId-1", Variation = "test-variation-1" }
                }
            },
            i => i.Index("variants"));

            // System.Console.WriteLine(indexResponse);
            // System.Console.WriteLine(indexResponse.Id);
            // System.Console.WriteLine(indexResponse.OriginalException);
            // System.Console.WriteLine(indexResponse.Result);
            // System.Console.WriteLine(indexResponse.ServerError);

        }

        [Fact]
        public void CanAddAndRemoveVariantWithSample()
        {        
            var username = fixture.ElasticUsername;
            var password = fixture.ElasticPassword;

            var url = $"{fixture.VariantsGatewayUrl}{fixture.PublicFacingElasticPath}";

            // Create ES Client with Basic Authentication header
            var settings = new ConnectionSettings(new Uri(url))
                .BasicAuthentication(username, password)
                .DefaultIndex("variants")
# if DEBUG
                .EnableDebugMode()
# endif
                .ServerCertificateValidationCallback((sender, cert, chain, errors) => { return true; });

            var client = new ElasticClient(settings);

            var indexResponse = AddVariantWithSample(client);

            // Check results
            Assert.Equal(indexResponse.Id, "test-id");
            Assert.True(indexResponse.IsValid);


            // Wait
            Thread.Sleep(1000);
            

            // // Verify by retrieving test doc
            string searchQuery = $"?ids=test-id";
            string searchUrl = $"{fixture.VariantsGatewayUrl}{fixture.GetVariantsByVariantIdPath}{searchQuery}";
            
            using (HttpResponseMessage response = fixture.client.GetAsync(searchUrl).Result)
            {
                var responseContent = response.Content.ReadAsStringAsync().Result;

                Assert.Equal(response.StatusCode, HttpStatusCode.OK);

                var dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);
                
                Assert.True(dto != null);

                Assert.Equal(dto.Status, 200);
                Assert.Equal(dto.Message, "Success");

                System.Console.WriteLine(Newtonsoft.Json.JsonConvert.SerializeObject(dto));

                Assert.True(dto.Data
                    .First(d => d.VariantId == "test-id")
                    .Count == 1);
            }


            // Wait
            Thread.Sleep(1000);
            

            // Verify removal of doc by sampleId
            string removeQuery = $"?id=test-sampleId-1";
            string removeUrl = $"{fixture.VariantsGatewayUrl}{fixture.RemoveSampleIdPath}{removeQuery}";
            
            using (HttpResponseMessage response = fixture.client.GetAsync(removeUrl).Result)
            {
                var responseContent = response.Content.ReadAsStringAsync().Result;

                Assert.Equal(response.StatusCode, HttpStatusCode.OK);

                var dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);
                
                Assert.True(dto != null);

                System.Console.WriteLine(dto.Message);

                Assert.Equal(dto.Status, 200);
                //Assert.Equal(dto.Message, "Success");

                //Assert.True(dto.Data.FirstOrDefault(d => d.VariantId == "test-id") == null);
            }


            // Wait
            Thread.Sleep(3000);
            

            using (HttpResponseMessage response = fixture.client.GetAsync(searchUrl).Result)
            {
                var responseContent = response.Content.ReadAsStringAsync().Result;

                Assert.Equal(response.StatusCode, HttpStatusCode.OK);

                var dto = JsonConvert.DeserializeObject<VariantsResponseDTO>(responseContent);
                
                Assert.True(dto != null);

                Assert.Equal(dto.Status, 200);
                Assert.Equal(dto.Message, "Success");

                System.Console.WriteLine(Newtonsoft.Json.JsonConvert.SerializeObject(dto));

                Assert.True(dto.Data
                    .First(d => d.VariantId == "test-id")
                    .Count == 0);
            }
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
                string url = $"{fixture.VariantsGatewayUrl}{fixture.GetVariantsByVariantIdPath}{query}";
                
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
                string url = $"{fixture.VariantsGatewayUrl}{fixture.GetVariantsByVariantIdPath}{query}";
                
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