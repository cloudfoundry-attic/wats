using System;
using System.Web.Http;

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

        [Route("~/headers")]
        [HttpGet]
        public IHttpActionResult Headers()
        {
            return Ok(Request.Headers);
        }

        [Route("~/print/{output}")]
        [HttpGet]
        public IHttpActionResult Print(string output)
        {
            System.Console.WriteLine(output);
            return Ok(Request.Headers);
        }

        [Route("~/id")]
        [HttpGet]
        public IHttpActionResult Id()
        {
            const string uuid = "A123F285-26B4-45F1-8C31-816DC5F53ECF";
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
    }
}