using System;

using Microsoft.AspNetCore.Mvc.Filters;

using Gohan.Api.Services.Interfaces;

namespace Gohan.Api.Middleware
{
    public class MandateAuthorizationTokensAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            Console.WriteLine("Authz validation middleware hit!");
            
            var authzService = (IAuthorizationService)(context.HttpContext.RequestServices.GetService(typeof(IAuthorizationService)));
            
            if (authzService.IsEnabled())
            {
                authzService.EnsureAllRequiredHeadersArePresent(context.HttpContext.Request.Headers);

                Console.WriteLine("All required headers are present!");


                // TODO : retrieve list of valid "datasets" (or other permitted tokens to query on)
                // for the time being, simply validate users access permission as "permitted" or "denied"
                
                // TEMP
                string authnTokenHeader = "X-AUTHN-TOKEN";

                var recoveredAuthnToken = string.Empty;

                if (context.HttpContext.Request.Headers.TryGetValue(authnTokenHeader, out var traceValue))
                    recoveredAuthnToken = traceValue;

                authzService.EnsureRepositoryAccessPermittedForUser(recoveredAuthnToken);
            }
        }
    }
}
