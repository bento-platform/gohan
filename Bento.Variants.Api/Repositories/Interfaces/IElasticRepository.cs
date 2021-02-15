using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariantInPositionRange(double? chromosome, string variant, double? lowerBound, double? upperBound);

        Task<long> CountDocumentsContainingSampleIdInPositionRange(double? chromosome, string sampleId, double? lowerBound, double? upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariantId(double? chromosome, string variant, double? lowerBound, double? upperBound, int size = 100, string sortByPosition = null);

        Task<List<dynamic>> GetDocumentsContainingSampleId(double? chromosome, string sampleId, double? lowerBound, double? upperBound, int size = 100, string sortByPosition = null);
    }
}
