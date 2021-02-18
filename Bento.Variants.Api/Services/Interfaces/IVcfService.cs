using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Services.Interfaces
{
    public interface IVcfService
    {
        Task<string> SynthesizeSingleSampleIdVcf(string headerBlockStrng, List<dynamic> docs);
    }
}