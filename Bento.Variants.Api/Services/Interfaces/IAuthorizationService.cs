using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Bento.Variants.Api.Services.Interfaces
{
    public interface IAuthorizationService 
    {
        bool IsEnabled();
        string GetOpaUrl();
        List<string> GetRequiredHeaders();

        void EnsureAllRequiredHeadersArePresent(Microsoft.AspNetCore.Http.IHeaderDictionary headers);
        void EnsureRepositoryAccessPermittedForUser(string username);

        bool IsGlobalRepositoryAccessPermitted();
    }
}