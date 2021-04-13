using System;
using System.Collections.Generic;

namespace Bento.Variants.Api.Services.Interfaces
{
    public interface IAuthorizationService 
    {
        bool IsEnabled();
        string GetOpaUrl();
        List<string> GetRequiredHeaders();

        bool AllRequiredHeadersArePresent(Microsoft.AspNetCore.Http.IHeaderDictionary headers);
        bool IsGlobalRepositoryAccessPermitted();
    }
}