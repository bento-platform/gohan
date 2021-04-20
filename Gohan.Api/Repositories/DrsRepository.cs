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

        private string searchObjects = "/search";
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
            var getUrl = $"{Configuration["Drs:PrivateUrl"]}{searchObjects}{forwardedQueryString}";

            // call drs
            var result = await httpClient.GetAsync(getUrl);
            var jsonData = await result.Content.ReadAsStringAsync();
        
            return jsonData;
        }
    }
}