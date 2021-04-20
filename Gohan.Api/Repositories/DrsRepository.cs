using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;

using Gohan.Api.Exceptions;
using Gohan.Api.Repositories.Interfaces;

using Microsoft.Extensions.Configuration;

using Newtonsoft.Json;

namespace Gohan.Api.Repositories
{
    public class DrsRepository : IDrsRepository
    {
        private readonly IConfiguration Configuration;

        private string searchObjectsPath = "/search";
        private string publicIngestPath = "/public/ingest";

        private string getObjectPath = "/objects/{0}";
        private string downloadObjectPath = "/objects/{0}/download";
        private HttpClient httpClient;
        
        public DrsRepository(IConfiguration configuration)
        {
            Configuration = configuration;

            HttpClientHandler httpClientHandler = new HttpClientHandler();
            httpClient = new HttpClient(httpClientHandler, disposeHandler: false);
        }


        public async Task<string> GetObjectById(string objectId)        
        {
            var getUrl = $"{Configuration["Drs:PrivateUrl"]}{string.Format(getObjectPath, objectId)}";

            // call drs
            var result = await httpClient.GetAsync(getUrl);
            var jsonData = await result.Content.ReadAsStringAsync();
        
            return jsonData;
        }

        public async Task<byte[]> DownloadObjectById(string objectId)
        {
            var downloadUrl = $"{Configuration["Drs:PrivateUrl"]}{string.Format(downloadObjectPath, objectId)}";

            // call drs
            var result = await httpClient.GetAsync(downloadUrl);
            var bytesData = await result.Content.ReadAsByteArrayAsync();

            return bytesData;
        }

        public async Task<string> SearchObjectsByQueryString(string forwardedQueryString)        
        {
            var getUrl = $"{Configuration["Drs:PrivateUrl"]}{searchObjectsPath}{forwardedQueryString}";

            // call drs
            var result = await httpClient.GetAsync(getUrl);
            var jsonData = await result.Content.ReadAsStringAsync();
        
            return jsonData;
        }

        public async Task<string> PublicIngestFile(byte[] fileBytes, string filename)
        {
            using (var content = new MultipartFormDataContent())
            {
                // setup file bytes to be uploaded to drs
                var byteArrayContent = new ByteArrayContent(fileBytes);
                content.Add(byteArrayContent, "file", filename);

                var ingestUrl = $"{Configuration["Drs:PrivateUrl"]}{publicIngestPath}";

                // call drs
                var result = httpClient.PostAsync(ingestUrl, content).Result;
                var jsonData = await result.Content.ReadAsStringAsync();

                return jsonData;
            }
        }
    }
}