using System;
using System.Collections.Generic;

using Bento.Variants.XCC.Models.Indexes;

namespace Bento.Variants.XCC.Models.DTOs
{
    public class VariantsResponseDTO
    {
        public int Status { get; set; } = 0;
        public string Message { get; set; } = null;
        public List<VariantResponseDataModel> Data { get; set; } = new List<VariantResponseDataModel>();
    }

    public class VariantResponseDataModel
    {
        public string VariantId { get; set; } = null;
        public string SampleId { get; set; } = null;
        public int? Count { get; set; } = null;
        public List<VariantIndex> Results { get; set; } = null;
    }
}