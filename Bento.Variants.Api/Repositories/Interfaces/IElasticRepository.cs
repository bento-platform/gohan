using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariant(double chromosome, string variant);
        Task<List<dynamic>> GetDocumentsContainingVariant(double chromosome, string variant, int rowCount = 100);
        Task<List<dynamic>> GetDocumentsInPositionRange(double chromosome, double lowerBound, double upperBound, int rowCount = 100);
        Task<List<dynamic>> GetDocumentsContainingVariantInPositionRange(double chromosome, string variant, double lowerBound, double upperBound, int rowCount = 100);
    }
}
