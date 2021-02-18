using System.Text;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.Extensions.Configuration;

using Bento.Variants.XCC;
using Bento.Variants.Api.Services.Interfaces;
using Bento.Variants.Api.Repositories.Interfaces;

namespace Bento.Variants.Api.Repositories
{
    public class VcfService : IVcfService
    {
        private readonly IConfiguration Configuration;
        private readonly IElasticRepository ElasticRepository;

        public VcfService(
            IConfiguration configuration,
            IElasticRepository elasticRepository)
        {
            this.Configuration = configuration;
            this.ElasticRepository = elasticRepository;
        }
        public async Task<string> SynthesizeSingleSampleIdVcf(string sampleId, string fileId, List<dynamic> docs)
        {
            var originalFile = await ElasticRepository.GetFileByFileId(fileId);
            
            string vcfHeaders = Utils.Unzip(Convert.FromBase64String(originalFile["compressedHeaderBlockBase64"]));
        
            // Synthetize VCF
            var keys = ((IDictionary<string, object>)docs.First()).Keys
                .Select(k => k?.ToUpper());


            StringBuilder synthesizedVcfBuilder = new StringBuilder();

            var orderedKeys = keys.Except(new List<string>(){"SAMPLES", "FILEID"})
                .OrderBy(d => Utils.VCFColumnOrder.IndexOf(d))
            .ToList();
            var keysHeaderString = string.Join("\t", orderedKeys);


            // Append sample IDs to header or
            var firstDoc = docs.First();
            foreach(var sample in firstDoc["samples"])
            {
                if (sample["sampleId"] == sampleId)
                    keysHeaderString += $"\t{sample["sampleId"]}";
            }

            // Add the headers
            synthesizedVcfBuilder.AppendLine(keysHeaderString);

            // Add the docs
            foreach(var doc in docs)
            {
                var orderedDoc = ((IDictionary<string, dynamic>)doc)
                    .OrderBy(b => 
                        Utils.VCFColumnOrder.FindIndex(a => a == b.Key.ToUpper()))
                    .ToDictionary(pair => pair.Key, pair => pair.Value);

                var _didRemove = orderedDoc.Remove("fileId");

                dynamic poppedSamples;
                orderedDoc.Remove("samples", out poppedSamples);

                // listB.OrderBy(b => listA.FindIndex(a => a.id == b.id));
                var docLine = "";
                foreach(var pair in orderedDoc)
                {
                    // Check type
                    if (pair.Value.GetType() == typeof(string) ||
                        pair.Value.GetType() == typeof(long) ||
                        pair.Value.GetType() == typeof(double) ||
                        pair.Value.GetType() == typeof(decimal))
                    {
                        long tempLong = 0;
                        string appenditure = "";
                        if (pair.Value.GetType() == typeof(long) &&
                            long.TryParse(pair.Value.ToString(), out tempLong) && tempLong < 0)
                        {
                            // replace -1 with period to better match original file
                            appenditure= ".";
                        }
                        else
                        {
                            appenditure = pair.Value.ToString();
                        }

                        docLine += $"\t{appenditure}";
                    }
                }

                // Appened variations in order at the end of the line
                foreach(var sample in poppedSamples)
                {
                    if (sample["sampleId"] == sampleId)
                        docLine += $"\t{sample["variation"]}";
                }
            
                
                synthesizedVcfBuilder.AppendLine(docLine);
            }
            
            return string.Join(string.Empty, 
                vcfHeaders, 
                synthesizedVcfBuilder.ToString()
            );
        }
    }
}
