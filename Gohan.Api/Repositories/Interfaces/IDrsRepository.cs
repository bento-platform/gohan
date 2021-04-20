using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Gohan.Api.Repositories.Interfaces
{
    public interface IDrsRepository
    {
        Task<string> GetObjectById(string objectId);
        Task<byte[]> DownloadObjectById(string objectId);
        Task<string> SearchObjectsByQueryString(string forwardedQueryString);

        Task<string> PublicIngestFile(byte[] fileBytes, string filename);
    }
}