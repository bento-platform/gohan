using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;


using Gohan.Api.Middleware;
using Gohan.Api.Repositories.Interfaces;
using Gohan.Api.Services.Interfaces;

using Gohan.XCC;

using Newtonsoft.Json;

namespace Gohan.Api.Controllers
{
    [Route("drs")]
    [MandateAuthorizationTokens]
    public class DrsController : Controller
    {
        private readonly IConfiguration Configuration;
        private readonly IDrsRepository DrsRepository;
        
        public DrsController(
            IDrsRepository drsRepository,
            IConfiguration configuration)
        {
            this.Configuration = configuration;
            this.DrsRepository = drsRepository;
        }

        [HttpGet]
        [Route("objects/{objectId}")]
        public async Task<dynamic> GetObjectsById([FromRoute] string objectId)
        {
            var jsonString = await DrsRepository.GetObjectById(objectId);
            return JsonConvert.DeserializeObject(jsonString); 
        }

        [HttpGet]
        [Route("objects/{objectId}/download")]
        public async Task<IActionResult> DownloadObjectsById([FromRoute] string objectId)
        {
            var objectBytes = await DrsRepository.DownloadObjectById(objectId);
            return File(objectBytes, "application/octet-stream");
        }

        [HttpGet]
        [Route("search")]
        public async Task<dynamic> SearchObjectsByQueryString()
        {
            var fullQueryString = Request.QueryString.Value;
            var jsonString = await DrsRepository.SearchObjectsByQueryString(fullQueryString);
            return JsonConvert.DeserializeObject(jsonString); 
        }

        [HttpPost]
        [Route("ingest")]
        public async Task<dynamic> IngestNewFiles(IFormFile file)
        {   
            if (file != null && file.Length > 0)
            {
                // TODO : implement quality checking
                
                // retrieved uploaded file bytes
                byte[] fileBytes = null;
                using (var memstream = new MemoryStream())
                {
                    await file.CopyToAsync(memstream);
                    fileBytes = memstream.ToArray();
                }

                var jsonString = await DrsRepository.PublicIngestFile(fileBytes, file.FileName);
                return JsonConvert.DeserializeObject(jsonString); 
            }
            else
                throw new Exception($"Empty file was uploaded! {file?.FileName} {file?.Length}");      
        }
    }
}
