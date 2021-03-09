using System;
using System.Collections.Generic;
using System.Threading.Tasks;

using Bento.Variants.XCC.Models.Indexes;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantOrSampleIdInPositionRange(long? chromosome,
            long? lowerBound, long? upperBound, 
            string variantId = null, string sampleId = null);
        
        Task<List<VariantIndex>> GetDocumentsContainingVariantOrSampleIdInPositionRange(
            long? chromosome,
            long? lowerBound, long? upperBound, 
            string variantId = null, string sampleId = null,  
            int size = 100, string sortByPosition = null,
            bool includeSamplesInResultSet = true);

        Task<dynamic> GetFileByFileId(string fileId);

        Task RemoveSampleFromVariantsBySampleId(string sampleId);
    }
}
