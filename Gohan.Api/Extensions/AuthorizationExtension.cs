using System;
using System.Linq;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

using Gohan.Api.Services;
using Gohan.Api.Services.Interfaces;

namespace Gohan.Api
{
    public static class AuthorizationExtension
    {
        public static void AddAuthorizationLayer(this IServiceCollection services, IConfiguration configuration)
        {
            var isEnabled = $"{configuration["Authorization:IsEnabled"]}";
            var agreedToDisabledAuthzRiskTerms = $"{configuration["Authorization:AgreedToDisabledAuthzRiskTerms"]}";
            var oidcJwksUrl = $"{configuration["Authentication:OidcPublicJwksUrl"]}";
            var opaUrl = $"{configuration["Authorization:OpaUrl"]}";
            var reqHeaders = $"{configuration["Authorization:RequiredHeadersCommaSep"]}".Split(",").ToList();

            if (string.IsNullOrEmpty(isEnabled) ||  string.IsNullOrEmpty(opaUrl) ||  reqHeaders.Any(h => string.IsNullOrEmpty(h)) ||
                isEnabled.Contains("not-set")   ||  opaUrl.Contains("not-set")   ||  reqHeaders.Any(h => h.Contains("not-set")))
            {
                throw new Exception($"STARTUP ERROR: Invalid Authorization configuration! -- Aborting");
            }

#if DEBUG == false
            if(Boolean.Parse(isEnabled) == false && Boolean.Parse(agreedToDisabledAuthzRiskTerms) == false)
            {
                throw new Exception($"STARTUP ERROR: Data Access Authorization is disabled and Gohan Api is not running in development mode!\n" +
                                     "The `IsEnabled` toggle switch is intended to be used in development mode only for ease of use.\n" +
                                     "If you know what you are doing, and this was configured intentionally, please note that all " +
                                     "data will be exposed publically and will be availble to anyone without discrimination!\n" +
                                     "If you agree to these terms, please set the `AgreedToDisabledAuthzRiskTerms` configuration to `true`. " +
                                     "Otherwise, please enable Data Access Authorization by setting the `IsEnabled` configuration to `true` and redeploy! -- Aborting");
            }
#endif

            var authzConfig = new AuthorizationService(Boolean.Parse(isEnabled), oidcJwksUrl, opaUrl, reqHeaders);
            services.AddSingleton<IAuthorizationService>(authzConfig);            
        }
    }
}
