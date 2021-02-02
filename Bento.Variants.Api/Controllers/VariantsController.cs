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
        public IActionResult CountVariantCounts([FromQuery] string labels) //, [FromQuery] int rowCount = 100)
        {
            if (labels == null)
            {
                return Json(new 
                {
                    Error = "Missing variant labels!" 
                });
            }

            try
            {
                Dictionary<string,long> countResults = new Dictionary<string, long>();

                var variantsList = labels.Split(",");
                
                // TODO: optimize - make 1 repo call with all labels at once
                Parallel.ForEach(variantsList, variant =>
                {
                    var count = ElasticRepository.CountDocumentsContainingVariant(variant).Result;
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
                    message = "Failed to get : " + ex.Message
                });
            }
        }
    }
}


