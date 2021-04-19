using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;

using Gohan.Api.Exceptions;
using Gohan.Api.Services.Interfaces;

using Newtonsoft.Json;

namespace Gohan.Api.Services
{
    public class AuthorizationService : IAuthorizationService
    {
        private bool isEnabled = false;
        private string oidcJwksUrl = string.Empty;
        private string opaUrl = string.Empty;
        private List<string> requiredHeaders = new List<string>();

        private HttpClient httpClient;
        
        public AuthorizationService(bool _isEnabled , string _oidcJwksUrl, string _opaUrl, List<string> _requiredHeaders)
        {
            isEnabled = _isEnabled;
            oidcJwksUrl = _oidcJwksUrl;
            opaUrl = _opaUrl;
            requiredHeaders = _requiredHeaders;

            HttpClientHandler httpClientHandler = new HttpClientHandler();
#if DEBUG
            httpClientHandler.AllowAutoRedirect = false;
            httpClientHandler.ServerCertificateCustomValidationCallback = (message, cert, chain, errors) => { return true; };
#endif
            httpClient = new HttpClient(httpClientHandler, disposeHandler: false);
        }

        public bool IsEnabled() { 
            
            Console.WriteLine($"Authz service is {(isEnabled ? "enabled" : "disabled")}!");
            return isEnabled;
        }

        public string GetOidcJwksUrl() 
        {
            return oidcJwksUrl;
        }
        public string GetOpaUrl() 
        {
            return opaUrl;
        }
        
        public List<string> GetRequiredHeaders() 
        {
            return requiredHeaders;
        }

        public void EnsureAllRequiredHeadersArePresent(Microsoft.AspNetCore.Http.IHeaderDictionary requiredHeaders) 
        {
            // Ensure presence of necessary custom headers
            GetRequiredHeaders().ForEach(reqHeader => 
            {
                var expectedValue = string.Empty;

                if (requiredHeaders.TryGetValue(reqHeader, out var traceValue))
                    expectedValue = traceValue;

                if (string.IsNullOrEmpty(expectedValue))
                {
                    throw new MissingRequiredHeadersException(reqHeader);
                }
            });
        }

        public void EnsureRepositoryAccessPermittedForUser(string authnToken)
        {
            if(IsEnabled())
            {
                // Authentication
                
                // fetch authN public JWKS from idp
                // TODO: cache result to minimize bandwidth usage
                var authnJwks = (httpClient.GetAsync(GetOidcJwksUrl()).Result).Content.ReadAsStringAsync().Result;


                // Authorization

                bool? isAccessPermitted = false;

                object inputJson = new {
                    input = new {
                        authN_token = authnToken,
                        authN_jwks = authnJwks
                    }
                };

                string input = JsonConvert.SerializeObject(inputJson);

                using (var content = new StringContent(input, Encoding.UTF8))
                {
                    // call opa
                    var result = httpClient.PostAsync(GetOpaUrl(), content).Result;

                    // TODO : type safety (remove dynamic, add a class)
                    var data = Newtonsoft.Json.JsonConvert.DeserializeObject<dynamic>(result.Content.ReadAsStringAsync().Result);

                    isAccessPermitted = data.result;
                    
                    Console.WriteLine($"Access permitted ? {isAccessPermitted}");
                }

                if (isAccessPermitted == null || isAccessPermitted == false)
                {
                    if (isAccessPermitted == null)
                        Console.WriteLine("INTERNAL ERROR : isAccessPermitted is null, all access attempts will be denied -- check the authz url configuration!");
                    
                    throw new DataAccessDeniedException();
                }
            }
        }

        public bool IsGlobalRepositoryAccessPermitted()
        {
            if(IsEnabled())
            {
                Console.WriteLine("Global repository access granted!");
                return true;
            }

            return false;
        }
    }
}