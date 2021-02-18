using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

using Bento.Variants.Api.Services.Interfaces;

using Bento.Variants.Api.Repositories.Interfaces;

namespace Bento.Variants.Api.Controllers
{
    [Route("vcfs")]
    public class VcfController : Controller
    {

        private readonly IConfiguration Configuration;
        private readonly IElasticRepository ElasticRepository;
        private readonly IVcfService VcfService;
        

        public VcfController(
            IVcfService vcfService,
            IElasticRepository elasticRepository,
            IConfiguration configuration)
        {
            this.Configuration = configuration;
            this.ElasticRepository = elasticRepository;
            this.VcfService = vcfService;
        }

        [HttpGet]
        [Route("get/by/sampleId")]
        public async Task<IActionResult> GetVariantsBySampleIds(
            [FromQuery] double chromosome, 
            [FromQuery] string id, 
            [FromQuery] double? lowerBound,
            [FromQuery] double? upperBound,
            [FromQuery] int size = 100)
        {
            if (string.IsNullOrEmpty(id))
            {
                string message = "missing sample ID!";

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

            string sortByPosition = "asc";

            try
            {
                Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

                var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                    null, id, 
                    lowerBound, upperBound, 
                    size, sortByPosition
                ).Result;

                string fileId = docs.First()["fileId"];

                var recombinedVcfFile = await this.VcfService.SynthesizeSingleSampleIdVcf(fileId, docs);
                
                return Content(recombinedVcfFile, "application/octet-stream"); 
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
    }
}
