using System;

using Microsoft.AspNetCore.Mvc.Filters;

namespace Bento.Variants.Api.Middleware
{
    public class MandateChromosomeAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            var chrom = context.HttpContext.Request.Query["chromosome"];
            int chromIntOut = 0;

            if (chrom.Count == 0 ||                             // Missing chromosome parameter
                !Int32.TryParse(chrom[0], out chromIntOut) ||   // Invalid chromosome data type
                chromIntOut <= 0)                               // Invalid chromosome value
            {           
                string message = "missing chromosome!";
                Console.WriteLine(message);
                throw new Exception(message);
            }
        }
    }
}
