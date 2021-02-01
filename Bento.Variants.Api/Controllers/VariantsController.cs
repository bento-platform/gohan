using System;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
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
        [Route("get")]
        public async Task<IActionResult> GetVariantCounts([FromQuery] string variant) //, [FromQuery] int rowCount = 100)
        {
            if (variant == null)
            {
                return Json(new 
                    {
                        Error = "missing variant!" 
                    });
            }

            try
            {
                var count = await ElasticRepository.CountDocumentsContainingVariant(variant);

                
                return Json(new 
                {
                    Count = count,
                    //Documents = result
                });            
            }
            catch (System.Exception ex)
            {
                Console.WriteLine($"Oops! : {ex.Message}");
                
                return Json(new {
                    status=500,
                    message= "Failed to get : " + ex.Message});
            }
        }
    }
}


