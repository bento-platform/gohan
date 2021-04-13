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
            
            if(authzService.AllRequiredHeadersArePresent(context.HttpContext.Request.Headers))
            {
                Console.WriteLine("All required headers are present!");
            }

            // TODO : retrieve list of valid "datasets" (or other permitted tokens to query on)
            // for the time being, simply validate users access permission globally as "permitted" or "denied" 
            if(authzService.IsGlobalRepositoryAccessPermitted() == false)
            {
                throw new Exception("Access Denied!");
            }        
        }
    }
}
