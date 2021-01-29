using System;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using System.Web.Http;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using System.IO;
using System.Net.Http.Headers;
using Microsoft.AspNetCore.Identity;
using System.Collections.Generic;
using Microsoft.AspNetCore.Authorization;
using Microsoft.EntityFrameworkCore;
using System.Text;

using Bento.Variants.Api.Repositories.Interfaces;

namespace Bento.Variants.Api.Controllers
{
    [Route("variants")]
    public class VariantsController : Controller
    {
        private readonly IConfiguration Configuration;
        private readonly IElasticRepository ElasticRepository;

        public VariantsController(
            IElasticRepository elasticRepository,
            IConfiguration configuration)
        {
            this.Configuration = configuration;
            this.ElasticRepository = elasticRepository;
        }

        [HttpGet]
        [Route("get")]
        public async Task<IActionResult> SimulateGet([FromQuery] int rowCount = 100)
        {
            try
            {
                var esGet = ElasticRepository.SimulateElasticSearchGet();

                if (esGet)
                {
                    return Json(new 
                    {
                        DidGet = esGet 
                    });
                }
            }
            catch (System.Exception ex)
            {
                return Json(new {
                    status=500,
                    message= "Failed to get : " + ex.Message});
            }

            return Json(new 
            {
                DidGet = false
            });
        }
    }
}


