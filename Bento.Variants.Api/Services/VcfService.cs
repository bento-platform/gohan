using System.Text;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.Extensions.Configuration;

using Bento.Variants.Api.Models.DTOs;
using Bento.Variants.Api.Services.Interfaces;
using Bento.Variants.Api.Repositories.Interfaces;
using Bento.Variants.XCC;

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
            var firstDoc = docs.FirstOrDefault();
            if (firstDoc == null)
                throw new Exception("Invalid VCF - No variants!");


            var keys = ((IDictionary<string, object>)firstDoc).Keys
                .Select(k => k?.ToUpper());

            StringBuilder synthesizedVcfBuilder = new StringBuilder();

            // Order the keys according to the formalized sequence VCF columns arrive in,
            // and exclude the ad hoc "samples" and "fileid" fields
            var orderedKeys = keys.Except(new List<string>(){"SAMPLES", "FILEID"})
                .OrderBy(d => Utils.VCFColumnOrder.IndexOf(d))
                .ToList();
            
            // Create our the VCF column row
            var keysHeaderString = string.Join("\t", orderedKeys);


            // Append sample IDs to header or
            foreach(var sample in firstDoc["samples"])
            {
                if (sample["sampleId"] == sampleId)
                    keysHeaderString += $"\t{sample["sampleId"]}";
            }

            // Add the headers to create our first block of VCF text
            synthesizedVcfBuilder.AppendLine(keysHeaderString);

            // Add the docs
            foreach(var doc in docs)
            {
                // Order the  document keys according to the formalized sequence VCF columns arrive in
                var orderedDoc = ((IDictionary<string, dynamic>)doc)
                    .OrderBy(b => 
                        Utils.VCFColumnOrder.FindIndex(a => a == b.Key.ToUpper()))
                    .ToDictionary(pair => pair.Key, pair => pair.Value);


                // exclude the ad hoc "fileid" field
                var _didRemove = orderedDoc.Remove("fileId");

                // retrieve the "samples" for later use..
                dynamic poppedSamples;
                orderedDoc.Remove("samples", out poppedSamples);

                // Iterate over each key-value pair in the ordered document, 
                // and append the values appropriately to the line
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
            
                // Create the variant row
                synthesizedVcfBuilder.AppendLine(docLine);
            }
            
            // return the full VCF
            return string.Join(string.Empty, 
                vcfHeaders, 
                synthesizedVcfBuilder.ToString()
            );
        }
    }
}
