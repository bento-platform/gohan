using System;
using System.Text.Json;
using System.Threading.Tasks;

using Microsoft.AspNetCore.Http;

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
            catch (Exception error)
            {
                var response = context.Response;
                response.ContentType = "application/json";

                var dto = new VariantsResponseDTO();
                
                string message = $"{error.Message}";
#if DEBUG
                message += $"\n{error.StackTrace}";
#endif
                Console.WriteLine(message);
                
                dto.Status = 500;
                dto.Message = $"Error : {message}";

                var result = JsonSerializer.Serialize( // redundant? TODO:refactor
                    dto, new JsonSerializerOptions 
                    {
                        PropertyNamingPolicy = JsonNamingPolicy.CamelCase // lowercase keys
                    });
                await response.WriteAsync(result);
            }
        }
    }
}