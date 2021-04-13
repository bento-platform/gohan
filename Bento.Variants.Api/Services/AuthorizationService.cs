using System;
using System.Collections.Generic;
using Bento.Variants.Api.Services.Interfaces;

namespace Bento.Variants.Api.Services
{
    public class AuthorizationService : IAuthorizationService
    {
        private bool isEnabled = false;
        private string opaUrl = string.Empty;
        private List<string> requiredHeaders = new List<string>();
        
        public AuthorizationService(bool _isEnabled , string _opaUrl, List<string> _requiredHeaders)
        {
            isEnabled = _isEnabled;
            opaUrl = _opaUrl;
            requiredHeaders = _requiredHeaders;
        }

        public bool IsEnabled() { 
            
            Console.WriteLine($"Authz service is {(isEnabled ? "enabled" : "disabled")}!");
            return isEnabled;
        }

        public string GetOpaUrl() 
        {
            if(IsEnabled())
                return opaUrl;
            throw new Exception("Authorization is disabled! Enable it, reboot the service, and try again!");
        }
        
        public List<string> GetRequiredHeaders() 
        {
            if(IsEnabled())
                return requiredHeaders;
            throw new Exception("Authorization is disabled! Enable it, reboot the service, and try again!");
        }

        public bool AllRequiredHeadersArePresent(Microsoft.AspNetCore.Http.IHeaderDictionary requiredHeaders) 
        {
            if(IsEnabled())
            {
                // Ensure presence of necessary custom headers
                GetRequiredHeaders().ForEach(rh => 
                {
                    var customHeader = string.Empty;

                    if (requiredHeaders.TryGetValue(rh, out var traceValue))
                        customHeader = traceValue;

                    if (string.IsNullOrEmpty(customHeader))
                    {
                        string message = $"Authorization : Missing {rh} header!";
                        Console.WriteLine(message);
                        throw new Exception(message);
                    }
                });

                return true;
            }

            throw new Exception("Authorization is disabled! Enable it, reboot the service, and try again!");
        }

        public bool IsGlobalRepositoryAccessPermitted()
        {
            if(IsEnabled())
            {
                // TODO: call opa with tokens and return based on that response
                Console.WriteLine("Global repository access granted!");
                return true;
            }

            throw new Exception("Authorization is disabled! Enable it, reboot the service, and try again!");
        }
    }
}