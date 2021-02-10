using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantInPositionRange(double? chromosome, string variant, double? lowerBound, double? upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariantInPositionRange(double? chromosome, string variant, double? lowerBound, double? upperBound, int rowCount = 100);
    
        Task<List<dynamic>> GetDocumentsBySampleId(double? chromosome, string sampleId, int rowCount = 100);
    }
}
