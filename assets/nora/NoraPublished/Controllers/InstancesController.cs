using System;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Net.Http;
using System.Web.Http;
using System.Web.Http.Results;
using MySql.Data.MySqlClient;
using Newtonsoft.Json;
using Nora.helpers;

namespace nora.Controllers
{
    public class InstancesController : ApiController
    {
        private static Services services;

        static InstancesController()
        {
            var env = Environment.GetEnvironmentVariable("VCAP_SERVICES");
            if (env != null)
            {
                services = JsonConvert.DeserializeObject<Services>(env);
            }
        }

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

        [Route("~/curl/{host}/{port}")]
        [HttpGet]
        public IHttpActionResult Curl(string host, int port)
        {
            var req = WebRequest.Create("http://" + host + ":" + port);
            req.Timeout = 1000;
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

        [Route("~/env/{name}")]
        [HttpGet]
        public IHttpActionResult EnvName(string name)
        {
            return Ok(Environment.GetEnvironmentVariable(name));
        }

        [Route("~/users")]
        [HttpGet]
        public IHttpActionResult Users()
        {
            if (services.UserProvided.Count == 0)
            {
                var msg = new HttpResponseMessage();
                msg.StatusCode = HttpStatusCode.NotFound;
                msg.Content = new StringContent("No services found");
                return new ResponseMessageResult(msg);
            }

            var service = services.UserProvided[0];

            var creds = service.Credentials;
            var username = creds["username"];
            var password = creds["password"];
            var host = creds["host"];


            var connString = String.Format("server={0};uid={1};pwd={2};database=mysql", host, username, password);

            Console.WriteLine("Connecting to mysql using {0}", connString);

            var users = new List<string>();

            using (var conn = new MySqlConnection())
            {
                conn.ConnectionString = connString;
                conn.Open();
                using (var cmd = new MySqlCommand("select user from mysql.user where user <> ''", conn))
                {
                    using (var reader = cmd.ExecuteReader())
                    {
                        while (reader.Read())
                        {
                            var colIdx = reader.GetOrdinal("User");
                            users.Add(reader.GetString(colIdx));
                        }
                    }
                }
            }
            return Ok(users);
        }
    }
}