using System.Collections.Generic;

namespace Bento.Variants.XCC.Models.Indexes
{
    public class VariantIndex
    {
        //"CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO", "FORMAT"
        public int Chrom { get; set; }
        public int Pos { get; set; }
        public string Id { get; set; }
        public string Ref { get; set; }
        public string Alt { get; set; }
        public int Qual { get; set; }
        public string Filter { get; set; }
        public string Info { get; set; } 
        public string Format { get; set; }
        public string FileId { get; set; }

        public List<Sample> Samples { get; set; }
    }
}