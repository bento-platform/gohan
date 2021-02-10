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
        [Route("get/by/sampleIds")]
        public IActionResult BySampleIds(
            [FromQuery] double? chromosome, 
            [FromQuery] string sampleIds, 
            [FromQuery] int rowCount = 100)
        {
            if (string.IsNullOrEmpty(sampleIds))
            {
                string message = "missing sample ids!";

                Console.WriteLine(message);
                return Json(new 
                {
                    Error = message
                });
            } 
            try
            {
                Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

                var sampleIdList = sampleIds.Split(",");
            
                // TODO: optimize - make 1 repo call with all labels at once
                Parallel.ForEach(sampleIdList, sampleId =>
                {
                    var docs = ElasticRepository.GetDocumentsBySampleId(chromosome, sampleId, rowCount).Result;
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
        [Route("count")]
        public IActionResult CountVariants(
            [FromQuery] double? chromosome, 
            [FromQuery] string labels, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound,
            [FromQuery] int rowCount = 100)
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

                if (string.IsNullOrEmpty(labels))
                    labels = "*";

                var variantsList = labels.Split(",");
            
                // TODO: optimize - make 1 repo call with all labels at once
                Parallel.ForEach(variantsList, variant =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariantInPositionRange(chromosome, variant, lowerBound, upperBound).Result;
                    countResults[variant] = count;
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
        [Route("get")]
        public IActionResult GetVariantsInRange(
            [FromQuery] double? chromosome, 
            [FromQuery] string labels, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound,
            [FromQuery] int rowCount = 100)
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

                if (string.IsNullOrEmpty(labels))
                    labels = "*";
                
                var variantsList = labels.Split(",");
                
                // TODO: optimize - make 1 repo call with all labels at once
                Parallel.ForEach(variantsList, variant =>
                {
                    var docs = ElasticRepository.GetDocumentsContainingVariantInPositionRange(chromosome, variant, lowerBound, upperBound, rowCount).Result;
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
    }
}
