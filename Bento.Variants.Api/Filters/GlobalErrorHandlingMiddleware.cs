using System;
using System.Text.Json;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Http;

using Bento.Variants.Api.Exceptions;
using Bento.Variants.XCC.Models.DTOs;

namespace Bento.Variants.Api.Middleware
{
    public class GlobalErrorHandlingMiddleware
    {
        private readonly RequestDelegate _next;

        public GlobalErrorHandlingMiddleware(RequestDelegate next)
        {
            _next = next;
        }

        public async Task Invoke(HttpContext context)
        {
            try
            {
                await _next(context);
            }
            catch (Exception exception)
            {
                // Prepare response structure
                var response = context.Response;
                response.ContentType = "application/json";

                var dto = new VariantsResponseDTO();

                string message = $"{exception.Message}";
#if DEBUG
                message += $"\n{exception.StackTrace}";
#endif
                Console.WriteLine(message);
            

                // Determine "api status code"
                if (exception.GetType() == typeof(MissingRequiredHeadersException))
                {
                    dto.Status = 400;
                }
                else if(exception.GetType() == typeof(DataAccessDeniedException))
                {
                    dto.Status = 401;
                }
                else
                {
                    dto.Status = 500;
                }


                dto.Message = $"Error : {message}";
                
                await response.WriteAsync(
                    JsonSerializer.Serialize(dto, new JsonSerializerOptions  // redundant? TODO:refactor
                    {
                        PropertyNamingPolicy = JsonNamingPolicy.CamelCase // lowercase keys
                    }));
            }
        }
    }
}