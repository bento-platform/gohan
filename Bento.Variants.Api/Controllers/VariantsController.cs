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
        [Route("count")]
        public IActionResult CountVariants(
            [FromQuery] double chromosome, 
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

                var variantsList = labels.Split(",");
                
                // TODO: optimize - make 1 repo call with all labels at once
                Parallel.ForEach(variantsList, variant =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariant(chromosome, variant).Result;
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
            [FromQuery] double chromosome, 
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

                if (!string.IsNullOrEmpty(labels))
                {
                    var variantsList = labels.Split(",");
                    
                    // TODO: optimize - make 1 repo call with all labels at once
                    Parallel.ForEach(variantsList, variant =>
                    {
                        if (lowerBound != null && upperBound != null)
                        {
                            var docs = ElasticRepository.GetDocumentsContainingVariantInPositionRange(chromosome, variant, lowerBound ?? 0, upperBound ?? 0, rowCount).Result;
                            docResults[variant] = docs;
                        }
                        else
                        {
                            var docs = ElasticRepository.GetDocumentsContainingVariant(chromosome, variant, rowCount).Result;
                            docResults[variant] = docs;
                        }
                    });

                    return Json(new
                    {
                        Count = docResults.Count,
                        Data = docResults
                    });            
                }
                else
                {
                    var docs = ElasticRepository.GetDocumentsInPositionRange(chromosome, lowerBound ?? 0, upperBound ?? 0, rowCount).Result;
                    return Json(new 
                    {
                        Count = docs.Count,
                        Data = docs
                    });            
                }
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




