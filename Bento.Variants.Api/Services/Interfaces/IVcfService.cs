using System.Collections.Generic;
using System.Threading.Tasks;

using Bento.Variants.XCC.Models.Indexes;

namespace Bento.Variants.Api.Services.Interfaces
{
    public interface IVcfService
    {
        Task<string> SynthesizeSingleSampleIdVcf(string sampleId, string fileId, List<VariantIndex> docs);
    }
}