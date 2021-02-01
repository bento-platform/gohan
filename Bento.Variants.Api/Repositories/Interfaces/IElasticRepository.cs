using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Repositories.Interfaces
{
    public interface IElasticRepository
    {
        Task<long> CountDocumentsContainingVariant(string variant);
        Task<List<dynamic>> GetDocumentsContainingVariant(string variant, int rowCount = 100);
    }
}
