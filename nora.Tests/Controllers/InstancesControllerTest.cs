using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Threading;
using System.Web.Http;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using nora;
using nora.Controllers;

namespace nora.Tests.Controllers
{
    [TestClass]
    public class InstancesControllerTest
    {
        InstancesController instancesController;

        [TestInitialize()]
        public void Startup()
        {
            instancesController = new InstancesController
            {
                Request = new HttpRequestMessage(HttpMethod.Get, "http://example.com"),
                Configuration = new HttpConfiguration()
            };
        }
        [TestMethod]
        public void Root()
        {
            var response = instancesController.Root();
            string resp;
            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            Assert.AreEqual("hello i am nora running on http://example.com/", resp);
        }

        [TestMethod]
        public void GetById()
        {
            var instanceGuid = Guid.NewGuid().ToString();
            Environment.SetEnvironmentVariable("INSTANCE_GUID", instanceGuid);
            var response = instancesController.Id();
            string resp;
            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            Assert.AreEqual(instanceGuid, resp);
        }

        [TestMethod]
        public void Env()
        {
            var response = instancesController.Env();
            Hashtable resp;
            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            CollectionAssert.AreEqual(Environment.GetEnvironmentVariables(), resp);
        }

        [TestMethod]
        public void EnvName()
        {
            Environment.SetEnvironmentVariable("FRED", "JANE");

            var response = instancesController.EnvName("FRED");
            string resp;
            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            Assert.AreEqual("JANE", resp);
        }

        [TestMethod]
        public void ReachableTcpIpPort()
        {
            var response = instancesController.Connect("8.8.8.8", 53);
            var json = response.ExecuteAsync(new CancellationToken()).Result.Content.ReadAsStringAsync();
            json.Wait();

            Assert.AreEqual("{\"stdout\":\"Successful TCP connection to 8.8.8.8:53\",\"stderr\":\"\",\"return_code\":0}", json.Result);
        }

        [TestMethod]
        public void UnreachableIpPort()
        {
            var response = instancesController.Connect("127.0.0.1", 20);
            var json = response.ExecuteAsync(new CancellationToken()).Result.Content.ReadAsStringAsync();
            json.Wait();
            Assert.AreEqual("{\"stdout\":\"\",\"stderr\":\"Unable to make TCP connection to 127.0.0.1:20\",\"return_code\":1}", json.Result);

        }

        [TestMethod]
        public void IpIsMissing()
        {
            var response = instancesController.Connect(null, 53);
            var json = response.ExecuteAsync(new CancellationToken()).Result.Content.ReadAsStringAsync();
            json.Wait();
            Assert.IsTrue(json.Result.Contains("\"return_code\":2"));
        }

        [TestMethod]
        public void PortIsInvalid()
        {
            var response = instancesController.Connect("127.0.0.1", IPEndPoint.MinPort - 1);
            var json = response.ExecuteAsync(new CancellationToken()).Result.Content.ReadAsStringAsync();
            json.Wait();
            Assert.IsTrue(json.Result.Contains("\"return_code\":2"));

            response = instancesController.Connect("127.0.0.1", IPEndPoint.MaxPort + 1);
            json = response.ExecuteAsync(new CancellationToken()).Result.Content.ReadAsStringAsync();
            json.Wait();
            Assert.IsTrue(json.Result.Contains("\"return_code\":2"));
        }

        [TestMethod]
        public void HealthCheck()
        {
            var response = instancesController.Healthcheck();
            string resp;
            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            Assert.AreEqual("Healthcheck passed", resp);
        }

        [TestMethod]
        public void RedirectTo()
        {
            var response = instancesController.RedirectTo("healthcheck");
            string resp;

            response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
            Assert.AreEqual("http://example.com/healthcheck", response.Location.ToString());
        }

        [TestMethod]
        public void PrintOutput()
        {

            var originalOut = Console.Out;

            using (StringWriter sw = new StringWriter())
            {
                Console.SetOut(sw);

                var response = instancesController.Print("output");
                string resp;

                response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                Assert.AreEqual(sw.ToString(), "output" + Environment.NewLine);
            }

            Console.SetOut(originalOut);
        }

        [TestMethod]
        public void PrintError()
        {

            var originalErr = Console.Error;

            using (StringWriter sw = new StringWriter())
            {
                Console.SetError(sw);

                var response = instancesController.PrintErr("error");
                string resp;

                response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                Assert.AreEqual(sw.ToString(), "error" + Environment.NewLine);
            }

            Console.SetError(originalErr);
        }

        [TestMethod]
        public void LogSpew()
        {

            var originalOut = Console.Out;

            using (StringWriter sw = new StringWriter())
            {
                Console.SetOut(sw);

                var response = instancesController.LogSpew(2);
                string resp;

                response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                Assert.IsTrue(sw.ToString().Contains(new string('1', 1024 * 1)));
            }

            Console.SetOut(originalOut);
        }
    }
}
