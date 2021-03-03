using System;
using System.Diagnostics;

using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Bento.Variants.Api.Filters
{
    public class MandateSampleIdSingularAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            if (context.HttpContext.Request.Query["id"].Count == 0)
            {           
                string message = "missing sample ID!";

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

    public class MandateSampleIdsPluralAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            if (context.HttpContext.Request.Query["ids"].Count == 0)
            {           
                string message = "missing sample IDs!";

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
