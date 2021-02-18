using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantOrSampleIdInPositionRange(double? chromosome, string variantId, string sampleId, double? lowerBound, double? upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariantOrSampleIdInPositionRange(
            double? chromosome, 
            string variantId, string sampleId, 
            double? lowerBound, double? upperBound, 
            int size = 100, string sortByPosition = null,
            bool includeSamplesInResultSet = true);

        Task<dynamic> GetFileByFileId(string fileId);
    }
}
