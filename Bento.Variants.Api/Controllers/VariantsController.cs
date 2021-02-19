using System;
using System.Collections.Concurrent;
using System.Linq;
using System.Collections.Generic;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

using Bento.Variants.Api.Repositories.Interfaces;
using Bento.Variants.Api.Models.DTOs;

using Bento.Variants.XCC;
// TODO: refactor paramter filtering code
namespace Bento.Variants.Api.Controllers
{
    [Route("variants")]
    public class VariantsController : Controller
    {
        private readonly IConfiguration Configuration;
        private readonly IElasticRepository ElasticRepository;

        public VariantsController(
            IElasticRepository elasticRepository,
            IConfiguration configuration)
        {
            this.Configuration = configuration;
            this.ElasticRepository = elasticRepository;
        }

        [HttpGet]
        [Route("get/by/sampleId")]
        public VariantsResponseDTO GetVariantsBySampleIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound,
            [FromQuery] int size = 100,
            [FromQuery] string sortByPosition = null,
            [FromQuery] bool includeSamplesInResultSet = true)
        {
            var response = new VariantsResponseDTO();

            if (string.IsNullOrEmpty(ids))
            {
                string message = "missing sample ids!";

                Console.WriteLine(message);

                response.Status = 500;
                response.Message = message;

                return response;
            
            } 

            if ((upperBound?.GetType() == typeof(long) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(long) && upperBound == null) ||
                upperBound < lowerBound)
            {
                response.Status = 500;
                response.Message = "Invalid lower and upper bounds!!";

                return response;
            }

            try
            {
                Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

                var sampleIdList = ids.Split(",");
            
                // TODO: optimize - make 1 repo call with all variantIds at once
                var tempResultsList = new ConcurrentBag<dynamic>();
                Parallel.ForEach(sampleIdList, sampleId =>
                {
                    var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                        null, sampleId, 
                        lowerBound, upperBound, 
                        size, sortByPosition,
                        includeSamplesInResultSet).Result;
                    results[sampleId] = docs;                    
                
                    tempResultsList.Add(new VariantResponseDataModel()
                    {
                        SampleId = sampleId,
                        Count = docs.Count,
                        Results = docs
                    });
                });

                response.Status = 200;
                response.Message = "Success";
                response.Data = tempResultsList.ToList();

                return response;
  
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                response.Status = 500;
                response.Message = "Failed to get : " + ex.Message;

                return response;
            }
        }

        [HttpGet]
        [Route("get/by/variantId")]
        public VariantsResponseDTO GetVariantsByVariantIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound,
            [FromQuery] int size = 100,
            [FromQuery] string sortByPosition = null,
            [FromQuery] bool includeSamplesInResultSet = false)
        {
            var response = new VariantsResponseDTO();

            if ((upperBound?.GetType() == typeof(long) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(long) && upperBound == null) ||
                upperBound < lowerBound)
            {
                response.Status = 500;
                response.Message = "Invalid lower and upper bounds!!";

                return response;
            }

            try
            {
                Dictionary<string,dynamic> docResults = new Dictionary<string, dynamic>();

                if (string.IsNullOrEmpty(ids))
                    ids = "*";
                
                var variantIdList = ids.Split(",");
                
                // TODO: optimize - make 1 repo call with all variantIds at once
                var tempResultsList = new ConcurrentBag<dynamic>();
                Parallel.ForEach(variantIdList, variant =>
                {
                    var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                        variant, null, 
                        lowerBound, upperBound, 
                        size, sortByPosition,
                        includeSamplesInResultSet).Result;
                
                    tempResultsList.Add(new VariantResponseDataModel()
                    {
                        VariantId = variant,
                        Count = docs.Count,
                        Results = docs
                    });
                });

                response.Status = 200;
                response.Message = "Success";
                response.Data = tempResultsList.ToList();

                return response;
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                response.Status = 500;
                response.Message = "Failed to get : " + ex.Message;

                return response;
            }
        }

        [HttpGet]
        [Route("count/by/variantId")]
        public VariantsResponseDTO CountVariantsByVariantIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound)
        {
            var response = new VariantsResponseDTO();

            if ((upperBound?.GetType() == typeof(long) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(long) && upperBound == null) ||
                upperBound < lowerBound)
            {
                response.Status = 500;
                response.Message = "Invalid lower and upper bounds!!";

                return response;
            }

            try
            {
                if (string.IsNullOrEmpty(ids))
                    ids = "*";

                var variantIdList = ids.Split(",");
            
                // TODO: optimize - make 1 repo call with all ids at once
                var tempResultsList = new ConcurrentBag<dynamic>();
                Parallel.ForEach(variantIdList, variantId =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, variantId, null, lowerBound, upperBound).Result;
                                    
                    tempResultsList.Add(new VariantResponseDataModel()
                    {
                        VariantId = variantId,
                        Count = (int)count,
                        Results = null
                    });
                });

                response.Status = 200;
                response.Message = "Success";
                response.Data = tempResultsList.ToList();

                return response;
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                response.Status = 500;
                response.Message = "Failed to count : " + ex.Message;

                return response;
            }
        }


        [HttpGet]
        [Route("count/by/sampleId")]
        public VariantsResponseDTO CountVariantsBySampleIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound)
        {
            var response = new VariantsResponseDTO();

            if ((upperBound?.GetType() == typeof(long) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(long) && upperBound == null) ||
                upperBound < lowerBound)
            {
                response.Status = 500;
                response.Message = "Invalid lower and upper bounds!!";

                return response;
            }

            try
            {
                if (string.IsNullOrEmpty(ids))
                    ids = "*";

                var sampleIdList = ids.Split(",");
            
                var tempResultsList = new ConcurrentBag<dynamic>();
                // TODO: optimize - make 1 repo call with all variantIds at once
                Parallel.ForEach(sampleIdList, sampleId =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, null, sampleId, lowerBound, upperBound).Result;

                    tempResultsList.Add(new VariantResponseDataModel()
                    {
                        SampleId = sampleId,
                        Count = (int)count,
                        Results = null
                    });
                });

                response.Status = 200;
                response.Message = "Success";
                response.Data = tempResultsList.ToList();

                return response;
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                response.Status = 500;
                response.Message = "Failed to count : " + ex.Message;

                return response;
            }
        }
    }
}
