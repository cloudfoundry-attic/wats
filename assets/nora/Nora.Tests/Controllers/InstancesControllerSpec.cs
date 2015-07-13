using System;
using System.Collections;
using System.Net.Http;
using System.Threading;
using System.Web.Http;
using nora.Controllers;
using NSpec;

namespace nora.Tests.Controllers
{
    internal class InstancesControllerSpec : nspec
    {
        private void describe_()
        {
            InstancesController instancesController = null;

            before = () =>
            {
                instancesController = new InstancesController
                {
                    Request = new HttpRequestMessage(HttpMethod.Get, "http://example.com"),
                    Configuration = new HttpConfiguration()
                };
            };

            describe["Get /"] = () =>
            {
                it["should return the hello message"] = () =>
                {
                    var response = instancesController.Root();
                    String resp = null;
                    response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                    resp.should_be("hello i am nora running on http://example.com/");
                };
            };

            describe["GET /id"] = () =>
            {
                it["should get the instance id from the INSTANCE_GUID"] = () =>
                {
                    var instanceGuid = Guid.NewGuid().ToString();
                    Environment.SetEnvironmentVariable("INSTANCE_GUID", instanceGuid);

                    var response = instancesController.Id();
                    String resp = null;
                    response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                    resp.should_be(instanceGuid);
                };
            }; 

            describe["Get /env"] = () =>
            {
                it["should return a list of ENV VARS"] = () =>
                {
                    var response = instancesController.Env();
                    Hashtable resp = null;
                    response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);
                    resp.should_be(Environment.GetEnvironmentVariables());
                };
            };

            describe["Get /env/:name"] = () =>
            {
                it["should return the desired named ENV VAR"] = () =>
                {
                    Environment.SetEnvironmentVariable("FRED", "JANE");

                    var response = instancesController.EnvName("FRED");
                    String resp = null;
                    response.ExecuteAsync(new CancellationToken()).Result.TryGetContentValue(out resp);

                    resp.should_be("JANE");
                };
            };
        }
    }
}