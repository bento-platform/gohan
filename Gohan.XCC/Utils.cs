using System.Collections.Generic;
using System.IO;
using System.IO.Compression;
using System.Text;

namespace Gohan.XCC
{
    public static class Utils
    {
        public static List<string> VCFColumnOrder = new List<string> ()
        {
            "CHROM", "POS", "ID", "REF", "ALT", "QUAL", "FILTER", "INFO", "FORMAT"
        };
 
        public static void CopyTo(Stream src, Stream dest)
        {
            byte[] bytes = new byte[4096];

            int cnt;

            while ((cnt = src.Read(bytes, 0, bytes.Length)) != 0)
            {
                dest.Write(bytes, 0, cnt);
            }
        }

        public static byte[] Zip(string str)
        {
            var bytes = Encoding.UTF8.GetBytes(str);

            using (var msi = new MemoryStream(bytes))
            using (var mso = new MemoryStream())
            {
                using (var gs = new GZipStream(mso, CompressionMode.Compress))
                {
                    //msi.CopyTo(gs);
                    CopyTo(msi, gs);
                }

                return mso.ToArray();
            }
        }

        public static string Unzip(byte[] bytes)
        {
            using (var msi = new MemoryStream(bytes))
            using (var mso = new MemoryStream()) 
            {
                using (var gs = new GZipStream(msi, CompressionMode.Decompress))
                {
                    //gs.CopyTo(mso);
                    CopyTo(gs, mso);
                }

                return Encoding.UTF8.GetString(mso.ToArray());
            }
        }
    }

    /// <summary>
    /// Random string generators.
    /// </summary>
    public static class RandomUtil
    {
        /// <summary>
        /// Get random string of 11 characters.
        /// </summary>
        /// <returns>Random string.</returns>
        public static string GetRandomString()
        {
            string path = Path.GetRandomFileName();
            path = path.Replace(".", ""); // Remove period.
            return path;
        }
    }
}