using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariant(string chromosome, string variant);
        Task<long> CountDocumentsInPositionRange(string chromosome, double lowerBound, double upperBound);
        Task<long> CountDocumentsContainingVariantInPositionRange(string chromosome, string variant, double lowerBound, double upperBound);
        
        Task<List<dynamic>> GetDocumentsContainingVariant(string chromosome, string variant, int rowCount = 100);
        Task<List<dynamic>> GetDocumentsInPositionRange(string chromosome, double lowerBound, double upperBound, int rowCount = 100);
        Task<List<dynamic>> GetDocumentsContainingVariantInPositionRange(string chromosome, string variant, double lowerBound, double upperBound, int rowCount = 100);
    
        Task<List<dynamic>> GetDocumentsBySampleId(string sampleId, int rowCount = 100);
    }
}
