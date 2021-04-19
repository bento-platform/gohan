using System.Collections.Generic;
using System.Threading.Tasks;

using Gohan.XCC.Models.Indexes;

namespace Gohan.Api.Services.Interfaces
{
    public interface IVcfService
    {
        Task<string> SynthesizeSingleSampleIdVcf(string sampleId, string fileId, List<VariantIndex> docs);
    }
}