using System;
namespace Bento.Variants.Api.Exceptions
{
    public class DataAccessDeniedException : Exception
    {
        private static string message = "Authorization : Access denied for {0}!";

        public DataAccessDeniedException(string username): base(String.Format(message, username)) {}
    }
}