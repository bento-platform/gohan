using System;
using System.Diagnostics;

using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

namespace Bento.Variants.Api.Filters
{
    public class MandateCalibratedBoundsAttribute : ActionFilterAttribute
    {
        public override void OnActionExecuting(ActionExecutingContext context)
        {
            long? lowerBound = null;
            long? upperBound = null;

            if (context.HttpContext.Request.Query["lowerBound"].Count > 0)
            {
                long tmpLower;
                bool didParse = long.TryParse(context.HttpContext.Request.Query["lowerBound"][0], out tmpLower);

                if (didParse) lowerBound = tmpLower;
            }

            if (context.HttpContext.Request.Query["upperBound"].Count > 0)
            {
                long tmpUpper;
                bool didParse = long.TryParse(context.HttpContext.Request.Query["upperBound"][0], out tmpUpper);

                if (didParse) upperBound = tmpUpper;
            }
          
            // Filter query parameters
            if ((upperBound?.GetType() == typeof(long) && lowerBound == null) ||
                (lowerBound?.GetType() == typeof(long) && upperBound == null) ||
                upperBound < lowerBound)
            {
                string message = "Invalid lower and upper bounds!!";

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
