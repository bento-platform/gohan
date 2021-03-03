using System;
using System.Diagnostics;

using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Bento.Variants.Api.Filters
{
    public class MandateChromosomeAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            if (context.HttpContext.Request.Query["chromosome"].Count == 0)
            {           
                string message = "missing chromosome!";

                Console.WriteLine(message);

                var result = new JsonResult(new 
                {
                    status = 500,
                    message = message
                });

                context.Result = result;
                return;
            }

        }
    }
}
