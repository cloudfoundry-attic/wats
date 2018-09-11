using nora.Helpers;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.IO.MemoryMappedFiles;
using System.Net;
using System.Net.Sockets;
using System.Web;
using System.Web.Http;
using System.Web.Http.Results;

namespace nora.Controllers
{
    public class InstancesController : ApiController
    {
        [Route("~/")]
        [HttpGet]
        public IHttpActionResult Root()
        {
            return Ok(String.Format("hello i am nora running on {0}", Request.RequestUri.AbsoluteUri));
        }

        [Route("~/id")]
        [HttpGet]
        public IHttpActionResult Id()
        {
            var uuid = Environment.GetEnvironmentVariable("INSTANCE_GUID");
            return Ok(uuid);
        }


        [Route("~/env")]
        [HttpGet]
        public IHttpActionResult Env()
        {
            return Ok(Environment.GetEnvironmentVariables());
        }

        [Route("~/env/{name}")]
        [HttpGet]
        public IHttpActionResult EnvName(string name)
        {
            return Ok(Environment.GetEnvironmentVariable(name));
        }


        [Route("~/connect/{host}/{port}")]
        [HttpGet]
        public IHttpActionResult Connect(string host, int port)
        {
            string stdout = "", stderr = "";
            int return_code = 0;
            TcpClient client = new TcpClient();
            try
            {
                client.Connect(host, port);
                return_code = 0;
                stdout = string.Format("Successful TCP connection to {0}:{1}", host, port);
            }
            catch (SocketException)
            {
                stderr = string.Format("Unable to make TCP connection to {0}:{1}", host, port);
                return_code = 1;
            }
            catch (Exception e)
            {
                stderr = e.Message;
                return_code = 2;
            }

            return Json(new
            {
                stdout = stdout,
                stderr = stderr,
                return_code = return_code
            });
        }

        [Route("~/healthcheck")]
        [HttpGet]
        public IHttpActionResult Healthcheck()
        {
            return Ok("Healthcheck passed");
        }
        
        [Route("~/redirect/{path}")]
        [HttpGet]
        public RedirectResult RedirectTo(string path)
        {
            var builder = new UriBuilder(Url.Request.RequestUri.DnsSafeHost)
            {
                Path = path,
            };
            return Redirect(builder.ToString());
        }

        [Route("~/print/{output}")]
        [HttpGet]
        public IHttpActionResult Print(string output)
        {
            System.Console.WriteLine(output);
            return Ok(Request.Headers);
        }

        [Route("~/print_err/{output}")]
        [HttpGet]
        public IHttpActionResult PrintErr(string output)
        {
            Console.Error.WriteLine(output);
            return Ok(Request.Headers);
        }

        [Route("~/curl/{host}/{port}")]
        [HttpGet]
        public IHttpActionResult Curl(string host, int port)
        {
            var req = WebRequest.Create("http://" + host + ":" + port);
            req.Timeout = 10000;
            try
            {
                var resp = (HttpWebResponse)req.GetResponse();
                return Json(new
                {
                    stdout = new StreamReader(resp.GetResponseStream()).ReadToEnd(),
                    return_code = 0,
                });
            }
            catch (WebException ex)
            {
                return Json(new
                {
                    stderr = ex.Message,
                    // ex.Response != null if the response status code wasn't a success,
                    // null if the operation timedout
                    return_code = ex.Response != null ? 0 : 1,
                });
            }
        }

        [Route("~/logspew/{kbytes}")]
        [HttpGet]
        public IHttpActionResult LogSpew(int kbytes)
        {
            var kb = new string('1', 1024);
            for (var i = 0; i < kbytes; i++)
            {
                Console.WriteLine(kb);
            }
            return Ok(String.Format("Just wrote {0} kbytes to the log", kbytes));
        }

        [Route("~/inaccessible_file")]
        [HttpPost]
        public IHttpActionResult InaccessibleFiles()
        {
            var result = Request.Content.ReadAsStringAsync().GetAwaiter().GetResult();
            return Ok(FileService.FileAccessStatus(result));
        }

        [Route("~/headers")]
        [HttpGet]
        public IHttpActionResult Headers()
        {
            return Ok(Request.Headers);
        }

        [Route("~/run")]
        [HttpPost]
        public IHttpActionResult Run()
        {
            var result = Request.Content.ReadAsStringAsync().GetAwaiter().GetResult();
            var path = HttpContext.Current.Request.MapPath(result);
            Process.Start(path);
            return Ok("Started: " + path);
        }

        [Route("~/existsonpath")]
        [HttpPost]
        public IHttpActionResult existsOnPath()
        {
            var result = Request.Content.ReadAsStringAsync().GetAwaiter().GetResult();
            Console.WriteLine(result);
            if (File.Exists(result))
                return Ok(Path.GetFullPath(result));

            var values = Environment.GetEnvironmentVariable("PATH");
            foreach (var path in values.Split(';'))
            {
                var fullPath = Path.Combine(path, result);
                if (File.Exists(fullPath))
                    return Ok(fullPath);
            }
            return Ok();
        }

        [Route("~/commitcharge")]
        [HttpGet]
        public IHttpActionResult GetCommitCharge()
        {
            var p = new PerformanceCounter("Memory", "Committed Bytes");
            return Ok(p.RawValue);
        }

        private static MemoryMappedFile MmapFile = null;

        [Route("~/mmapleak/{maxbytes}")]
        [HttpGet]
        public IHttpActionResult MmapLeakMax(long maxbytes)
        {
            if (MmapFile != null)
            {
                MmapFile.Dispose();
            }

            MmapFile = MemoryMappedFile.CreateNew(
                Guid.NewGuid().ToString(),
                maxbytes,
                MemoryMappedFileAccess.ReadWrite);

            return Ok();
        }

        [Route("~/exit")]
        [HttpGet]
        public IHttpActionResult Exit()
        {
            Process.GetCurrentProcess().Kill();
            return Ok();
        }

        private static List<IntPtr> _leakedPointers;
        [Route("~/leakmemory/{mb}")]
        [HttpGet]
        public IHttpActionResult Memory(int mb)
        {
            if (_leakedPointers == null)
                _leakedPointers = new List<IntPtr>();

            var bytes = mb * 1024 * 1024;
            _leakedPointers.Add(System.Runtime.InteropServices.Marshal.AllocHGlobal(bytes));
            return Ok();
        }
    }
}
