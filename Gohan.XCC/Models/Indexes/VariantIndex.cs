using System.Collections.Generic;

namespace Gohan.XCC.Models.Indexes
{
    public class VariantIndex
    {
        //"CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO", "FORMAT"
        public int Chrom { get; set; }
        public int Pos { get; set; }
        public string Id { get; set; }
        public List<string> Ref { get; set; }
        public List<string> Alt { get; set; }
        public int Qual { get; set; }
        public string Filter { get; set; }
        public List<dynamic> Info { get; set; } // TODO: type safe
        public string Format { get; set; }
        public string FileId { get; set; }

        public List<Sample> Samples { get; set; }
    }
}