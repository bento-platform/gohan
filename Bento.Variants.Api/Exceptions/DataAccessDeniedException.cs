using System;
namespace Bento.Variants.Api.Exceptions
{
    public class DataAccessDeniedException : Exception
    {
        private static string message = "Authorization : Access denied !";

        public DataAccessDeniedException(): base(message) {}
    }
}