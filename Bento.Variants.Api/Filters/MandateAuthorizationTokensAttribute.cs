using System;

using Microsoft.AspNetCore.Mvc.Filters;

using Bento.Variants.Api.Services.Interfaces;

namespace Bento.Variants.Api.Middleware
{
    public class MandateAuthorizationTokensAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            Console.WriteLine("Authz validation middleware hit!");
            
            var authzService = (IAuthorizationService)(context.HttpContext.RequestServices.GetService(typeof(IAuthorizationService)));
            
            authzService.EnsureAllRequiredHeadersArePresent(context.HttpContext.Request.Headers);

            Console.WriteLine("All required headers are present!");


            // TODO : retrieve list of valid "datasets" (or other permitted tokens to query on)
            // for the time being, simply validate users access permission as "permitted" or "denied"
            
            // TEMP
            string usernameHeader = "X-USERNAME";
            var recoveredUsername = string.Empty;

            if (context.HttpContext.Request.Headers.TryGetValue(usernameHeader, out var traceValue))
                recoveredUsername = traceValue;

            if (string.IsNullOrEmpty(recoveredUsername))
            {
                string message = $"Authorization : Missing {usernameHeader} header!";
                Console.WriteLine(message);
                throw new Exception(message);
            }

            authzService.EnsureRepositoryAccessPermittedForUser(recoveredUsername);
        }
    }
}
