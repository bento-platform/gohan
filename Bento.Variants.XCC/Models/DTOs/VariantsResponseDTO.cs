using System;
using System.Collections.Generic;

namespace Bento.Variants.XCC.Models.DTOs
{
    public class VariantsResponseDTO
    {
        public List<VariantResponseDataModel> Data { get; set; } = new List<VariantResponseDataModel>();
        public int Status { get; set; } = 0;
        public string Message { get; set; } = null;
    }

    public class VariantResponseDataModel
    {
        public string VariantId { get; set; } = null;
        public string SampleId { get; set; } = null;
        public int? Count { get; set; } = null;
        public List<dynamic> Results { get; set; } = null;
    }
}