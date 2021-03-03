using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

using Bento.Variants.Api.Middleware;
using Bento.Variants.Api.Repositories;
using Bento.Variants.Api.Repositories.Interfaces;

using Bento.Variants.Api.Services.Interfaces;

namespace Bento.Variants.Api
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        public void ConfigureServices(IServiceCollection services)
        {
            services.Configure<CookiePolicyOptions>(options =>
            {
                // This lambda determines whether user consent for non-essential cookies is needed for a given request.
                options.CheckConsentNeeded = context => true;
                options.MinimumSameSitePolicy = SameSiteMode.None;
            });


            // Logs + Elastic
            services.AddElasticSearch(Configuration);

            // MVC
            services.AddMvc().SetCompatibilityVersion(CompatibilityVersion.Version_3_0);
        
            // -- IoC configuration --
            ConfigureServiceIoC(services);
            ConfigureRepositoryIoC(services);

        }

        private void ConfigureServiceIoC(IServiceCollection services)
        {
            // -- Service Configuration --
            services.AddTransient<IVcfService, VcfService>();
            // -- - --
        }


        private void ConfigureRepositoryIoC(IServiceCollection services)
        {
            // -- Repository Configuration --
            services.AddTransient<IElasticRepository, ElasticRepository>();
            // -- - --
        }


        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(IApplicationBuilder app, IWebHostEnvironment env)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
            }
            else
            {
                app.UseExceptionHandler("/Home/Error");
                app.UseHsts();
            }

            app.UseHttpsRedirection();
            app.UseStaticFiles();
            app.UseCookiePolicy();

            // global error handler
            app.UseMiddleware<GlobalErrorHandlingMiddleware>();

            app.UseRouting();

            app.UseEndpoints(e => 
                e.MapControllerRoute(
                    name: "default",
                    pattern: "{controller=Home}/{action=Index}/{id?}")
            );
        }
    }
}
