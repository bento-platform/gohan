using System;
namespace Bento.Variants.Api.Exceptions
{
    public class MissingRequiredHeadersException : Exception
    {
        private static string message = "Authorization : Missing required {0} header!";

        public MissingRequiredHeadersException(string missingHeader): base(String.Format(message, missingHeader)) {}
    }
}