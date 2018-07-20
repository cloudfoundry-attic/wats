using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Security;
using System.Web;

namespace nora.Helpers
{
    public class FileService
    {
        public static string FileAccessStatus(string path)
        {
            try
            {
                Directory.EnumerateFiles(path);
                return "ACCESS_ALLOWED";
            }
            catch (UnauthorizedAccessException)
            {
                return "ACCESS_DENIED";
            }
            catch (SecurityException)
            {
                return "ACCESS_DENIED";
            }
            catch (Exception)
            {
                if (File.Exists(path))
                {
                    return "ACCESS_ALLOWED";
                }
                try
                {
                    var stream = File.OpenRead(path);
                    stream.Close();
                }
                catch (UnauthorizedAccessException)
                {
                    return "ACCESS_DENIED";
                }
                catch (FileNotFoundException)
                {
                    return "NOT_EXIST";
                }
                catch (Exception ex)
                {
                    return "EXCEPTION: " + ex.ToString();
                }
                return "ACCESS_ALLOWED";
            }
        }
    }
}