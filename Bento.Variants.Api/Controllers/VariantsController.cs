using System;
using System.Collections.Concurrent;
using System.Linq;
using System.Collections.Generic;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;

using Bento.Variants.Api.Middleware;
using Bento.Variants.Api.Repositories.Interfaces;

using Bento.Variants.XCC.Models.DTOs;

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
        [MandateSampleIdsPlural]
        [MandateCalibratedBounds]
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

            Dictionary<string, dynamic> results = new Dictionary<string, dynamic>();

            var sampleIdList = ids.Split(",");
        
            // TODO: optimize - make 1 repo call with all variantIds at once
            var tempResultsList = new ConcurrentBag<VariantResponseDataModel>();
            Parallel.ForEach(sampleIdList, sampleId =>
            {
                var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                    lowerBound, upperBound, 
                    variantId: null, sampleId: sampleId,
                    size: size, sortByPosition: sortByPosition,
                    includeSamplesInResultSet: includeSamplesInResultSet).Result;
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

        [HttpGet]
        [MandateCalibratedBounds]
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

            Dictionary<string,dynamic> docResults = new Dictionary<string, dynamic>();

            if (string.IsNullOrEmpty(ids))
                ids = "*";
            
            var variantIdList = ids.Split(",");
            
            // TODO: optimize - make 1 repo call with all variantIds at once
            var tempResultsList = new ConcurrentBag<VariantResponseDataModel>();
            Parallel.ForEach(variantIdList, variant =>
            {
                var docs = ElasticRepository.GetDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                    lowerBound, upperBound, 
                    variantId: variant, sampleId: null,
                    size: size, sortByPosition: sortByPosition,
                    includeSamplesInResultSet: includeSamplesInResultSet).Result;
            
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

        [HttpGet]
        [MandateCalibratedBounds]
        [Route("count/by/variantId")]
        public VariantsResponseDTO CountVariantsByVariantIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound)
        {
            var response = new VariantsResponseDTO();

            if (string.IsNullOrEmpty(ids))
                ids = "*";

            var variantIdList = ids.Split(",");
        
            // TODO: optimize - make 1 repo call with all ids at once
            var tempResultsList = new ConcurrentBag<VariantResponseDataModel>();
            Parallel.ForEach(variantIdList, variantId =>
            {
                var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                    lowerBound, upperBound,
                    variantId: variantId, sampleId: null).Result;
                                
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


        [HttpGet]
        [MandateCalibratedBounds]
        [Route("count/by/sampleId")]
        public VariantsResponseDTO CountVariantsBySampleIds(
            [FromQuery] long? chromosome, 
            [FromQuery] string ids, 
            [FromQuery] long? lowerBound,
            [FromQuery] long? upperBound)
        {
            var response = new VariantsResponseDTO();

            if (string.IsNullOrEmpty(ids))
                ids = "*";

            var sampleIdList = ids.Split(",");
        
            var tempResultsList = new ConcurrentBag<VariantResponseDataModel>();
            // TODO: optimize - make 1 repo call with all variantIds at once
            Parallel.ForEach(sampleIdList, sampleId =>
            {
                var count = ElasticRepository.CountDocumentsContainingVariantOrSampleIdInPositionRange(chromosome, 
                    lowerBound, upperBound,
                    variantId: null, sampleId: sampleId).Result;

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

        [HttpGet]
        [MandateSampleIdSingularAttribute]
        [Route("remove/sampleId")]
        public async Task<VariantsResponseDTO> RemoveSampleIds([FromQuery] string id)
        {
            var response = new VariantsResponseDTO();
            
            await ElasticRepository.RemoveSampleFromVariantsBySampleId(id);

            response.Status = 200;
            response.Message = "Removed Successfully";

            return response;
        }
    }
}
