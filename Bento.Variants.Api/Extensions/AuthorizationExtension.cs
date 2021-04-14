using System;
using System.Linq;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

using Bento.Variants.Api.Services;
using Bento.Variants.Api.Services.Interfaces;

namespace Bento.Variants.Api
{
    public static class AuthorizationExtension
    {
        public static void AddAuthorizationLayer(this IServiceCollection services, IConfiguration configuration)
        {
            var isEnabled = $"{configuration["Authorization:IsEnabled"]}";
            var oidcJwksUrl = $"{configuration["Authentication:OidcPublicJwksUrl"]}";
            var opaUrl = $"{configuration["Authorization:OpaUrl"]}";
            var reqHeaders = $"{configuration["Authorization:RequiredHeadersCommaSep"]}".Split(",").ToList();

            if (string.IsNullOrEmpty(isEnabled) ||  string.IsNullOrEmpty(opaUrl) ||  reqHeaders.Any(h => string.IsNullOrEmpty(h)) ||
                isEnabled.Contains("not-set")   ||  opaUrl.Contains("not-set")   ||  reqHeaders.Any(h => h.Contains("not-set")))
            {
                throw new Exception($"STARTUP ERROR: Invalid Authorization configuration! -- Aborting");
            }

#if DEBUG == false
            if(Boolean.Parse(isEnabled) == false)
            {
                throw new Exception($"STARTUP ERROR: Not running in development mode, but Data Access Authorization is disabled!" +
                                     "Please enable Data Access Authorization and redeploy! -- Aborting");
            }
#endif

            var authzConfig = new AuthorizationService(Boolean.Parse(isEnabled), oidcJwksUrl, opaUrl, reqHeaders);
            services.AddSingleton<IAuthorizationService>(authzConfig);            
        }
    }
}
