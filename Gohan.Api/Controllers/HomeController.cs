using System.Diagnostics;

using Microsoft.AspNetCore.Mvc;

using Gohan.Api.Models;

namespace Gohan.Api.Controllers
{
    public class HomeController : Controller
    {
        public string Index()
        {
            return "Gohan - A Genomic Variants API";
        }

        [ResponseCache(Duration = 0, Location = ResponseCacheLocation.None, NoStore = true)]
        public IActionResult Error()
        {
            return View(new ErrorViewModel { RequestId = Activity.Current?.Id ?? HttpContext.TraceIdentifier });
        }
    }
}
