using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Gohan.Api.Services.Interfaces
{
    public interface IAuthorizationService 
    {
        bool IsEnabled();
        string GetOpaUrl();
        List<string> GetRequiredHeaders();

        void EnsureAllRequiredHeadersArePresent(Microsoft.AspNetCore.Http.IHeaderDictionary headers);
        void EnsureRepositoryAccessPermittedForUser(string authnToken);

        bool IsGlobalRepositoryAccessPermitted();
    }
}