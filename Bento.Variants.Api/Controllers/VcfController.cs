using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

using Bento.Variants.Api.Middleware;
using Bento.Variants.Api.Repositories.Interfaces;
using Bento.Variants.Api.Services.Interfaces;

using Bento.Variants.XCC;

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
        [MandateChromosome]
        [MandateCalibratedBounds]
        [MandateSampleIdSingular]
        [Route("get/by/sampleId")]
        public async Task<IActionResult> GetVariantsBySampleIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string id, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound,
            [FromQuery] int size = 100)
        {
            // Force ascending sort order
            string sortByPosition = "asc";

            Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

            var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                lowerBound, upperBound, 
                variantId: null, sampleId: id,
                size: size, sortByPosition: sortByPosition
            ).Result;

            string fileId = docs.FirstOrDefault()?.FileId;

            if (fileId == null)
                throw new Exception("No VCF available!");

            var recombinedVcfFile = await this.VcfService.SynthesizeSingleSampleIdVcf(id, fileId, docs);
            
            return Content(recombinedVcfFile, "application/octet-stream");             
        }
    }
}
