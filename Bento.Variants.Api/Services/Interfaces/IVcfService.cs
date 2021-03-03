﻿using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Services.Interfaces
{
    public interface IVcfService
    {
        Task<string> SynthesizeSingleSampleIdVcf(string sampleId, string fileId, List<dynamic> docs);
    }
}