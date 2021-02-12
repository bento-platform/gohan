using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantInPositionRange(double? chromosome, string variant, double? lowerBound, double? upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariantId(double? chromosome, string variant, double? lowerBound, double? upperBound, int rowCount = 100);
    
        Task<List<dynamic>> GetDocumentsContainingSampleId(double? chromosome, string sampleId, double? lowerBound, double? upperBound, int rowCount = 100);
    }
}
