using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantOrSampleIdInPositionRange(long? chromosome, string variantId, string sampleId, long? lowerBound, long? upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariantOrSampleIdInPositionRange(
            long? chromosome, 
            string variantId, string sampleId, 
            long? lowerBound, long? upperBound, 
            int size = 100, string sortByPosition = null,
            bool includeSamplesInResultSet = true);

        Task<dynamic> GetFileByFileId(string fileId);
    }
}
