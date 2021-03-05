using System;
using System.Diagnostics;

using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

using Bento.Variants.XCC.Models.DTOs;

namespace Bento.Variants.Api.Middleware
{
    public class MandateSampleIdSingularAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            if (context.HttpContext.Request.Query["id"].Count == 0)
            {           
                string message = "missing sample ID!";

                Console.WriteLine(message);
                throw new Exception(message);
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
                throw new Exception(message);
            }
        }
    }
}
