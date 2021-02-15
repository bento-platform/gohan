using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

using Newtonsoft.Json;

using Bento.Variants.Api.Repositories.Interfaces;

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
        public IActionResult GetVariantsBySampleIds(
            [FromQuery] double? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound,
            [FromQuery] int size = 100,
            [FromQuery] string sortByPosition = null)
        {
            if (string.IsNullOrEmpty(ids))
            {
                string message = "missing sample ids!";

                Console.WriteLine(message);
                return Json(new 
                {
                    Error = message
                });
            } 

            if ((upperBound?.GetType() == typeof(double) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(double) && upperBound == null) ||
                upperBound < lowerBound)
            {
                return Json(new 
                {
                    Error = "Invalid lower and upper bounds!!" 
                });
            }

            try
            {
                Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

                var sampleIdList = ids.Split(",");
            
                // TODO: optimize - make 1 repo call with all variantIds at once
                Parallel.ForEach(sampleIdList, sampleId =>
                {
                    var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, null, sampleId, lowerBound, upperBound, size, sortByPosition).Result;
                    results[sampleId] = docs;                    
                });

                return Json(results);    
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                return Json(new 
                {
                    status = 500,
                    message = "Failed to get variants by sample ids : " + ex.Message
                });
            }
        }

        [HttpGet]
        [Route("get/by/variantId")]
        public IActionResult GetVariantsByVariantIds(
            [FromQuery] double? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound,
            [FromQuery] int size = 100,
            [FromQuery] string sortByPosition = null)
        {
            if ((upperBound?.GetType() == typeof(double) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(double) && upperBound == null) ||
                upperBound < lowerBound)
            {
                return Json(new 
                {
                    Error = "Invalid lower and upper bounds!!" 
                });
            }

            try
            {
                Dictionary<string,dynamic> docResults = new Dictionary<string, dynamic>();

                if (string.IsNullOrEmpty(ids))
                    ids = "*";
                
                var variantIdList = ids.Split(",");
                
                // TODO: optimize - make 1 repo call with all variantIds at once
                Parallel.ForEach(variantIdList, variant =>
                {
                    var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, variant, null, lowerBound, upperBound, size, sortByPosition).Result;
                    docResults[variant] = docs;
                });

                return Json(new
                {
                    Count = docResults.Count,
                    Data = docResults
                });
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                return Json(new 
                {
                    status = 500,
                    message = "Failed to get : " + ex.Message
                });
            }
        }

        [HttpGet]
        [Route("count/by/variantId")]
        public IActionResult CountVariantsByVariantIds(
            [FromQuery] double? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound)
        {
            if ((upperBound?.GetType() == typeof(double) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(double) && upperBound == null) ||
                upperBound < lowerBound)
            {
                return Json(new 
                {
                    Error = "Invalid lower and upper bounds!!" 
                });
            }

            try
            {
                Dictionary<string,long> countResults = new Dictionary<string, long>();

                if (string.IsNullOrEmpty(ids))
                    ids = "*";

                var variantIdList = ids.Split(",");
            
                // TODO: optimize - make 1 repo call with all ids at once
                Parallel.ForEach(variantIdList, variantId =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, variantId, null, lowerBound, upperBound).Result;
                    countResults[variantId] = count;
                });

                return Json(countResults);    
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                return Json(new 
                {
                    status = 500,
                    message = "Failed to count : " + ex.Message
                });
            }
        }


        [HttpGet]
        [Route("count/by/sampleId")]
        public IActionResult CountVariantsBySampleIds(
            [FromQuery] double? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound)
        {
            if ((upperBound?.GetType() == typeof(double) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(double) && upperBound == null) ||
                upperBound < lowerBound)
            {
                return Json(new 
                {
                    Error = "Invalid lower and upper bounds!!" 
                });
            }

            try
            {
                Dictionary<string,long> countResults = new Dictionary<string, long>();

                if (string.IsNullOrEmpty(ids))
                    ids = "*";

                var sampleIdList = ids.Split(",");
            
                // TODO: optimize - make 1 repo call with all variantIds at once
                Parallel.ForEach(sampleIdList, sampleId =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, null, sampleId, lowerBound, upperBound).Result;
                    countResults[sampleId] = count;
                });

                return Json(countResults);    
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                return Json(new 
                {
                    status = 500,
                    message = "Failed to count : " + ex.Message
                });
            }
        }
    }
}
