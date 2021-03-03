using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using System;
using System.Collections.Generic;
using System.Net;
using System.Text.Json;
using System.Threading.Tasks;

using Bento.Variants.Api.Models.DTOs;

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

                Console.WriteLine($"Oops! : {error.Message}");
                
                dto.Status = 500;
                dto.Message = "Error : " + error.Message;

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