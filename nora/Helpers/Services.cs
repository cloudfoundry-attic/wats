using Newtonsoft.Json;
using System;
using System.Collections.Generic;

namespace nora.Helpers
{
    public class Service
    {
        [JsonProperty("name")]
        public string Name { get; internal set; }
        [JsonProperty("label")]
        public string Label { get; internal set; }
        [JsonProperty("tags")]
        public List<string> Tags { get; internal set; }
        [JsonProperty("credentials")]
        public IDictionary<string, string> Credentials { get; internal set; }
    }


    public class Services
    {
        [JsonProperty("user-provided")]
        public List<Service> UserProvided { get; private set; }

        [JsonProperty("p-mysql")]
        public List<Service> PMySQL { get; private set; }

    }
    /*
    public class UsersFromService
    {
        private Services services;

        public UsersFromService()
        {
            var env = Environment.GetEnvironmentVariable("VCAP_SERVICES");
            if (env != null)
            {
                services = JsonConvert.DeserializeObject<Services>(env);
            }
        }
        public List<string> GetUserProvidedUsers()
        {
            return get(services.UserProvided[0]);
        }

        public List<string> GetPMysqlUsers()
        {
            return get(services.PMySQL[0]);
        }

        List<string> get(Service service)
        {
            var creds = service.Credentials;
            var username = creds["username"];
            var password = creds["password"];
            var host = creds.ContainsKey("host") ? creds["host"] : creds["hostname"];
            var dbname = creds.ContainsKey("name") ? creds["name"] : "mysql";
            var connString = String.Format("server={0};uid={1};pwd={2};database={3}", host, username, password, dbname);

            Console.WriteLine("Connecting to mysql using {0}", connString);

            var users = new List<string>();

            using (var conn = new MySqlConnection())
            {
                conn.ConnectionString = connString;
                conn.Open();

                new MySqlCommand(
                    "CREATE TABLE IF NOT EXISTS Hits(Id INT PRIMARY KEY AUTO_INCREMENT, CreatedAt DATETIME) ENGINE=INNODB;", conn)
                    .ExecuteNonQuery();

                new MySqlCommand(
                    "INSERT INTO Hits(CreatedAt)VALUES(now());", conn)
                    .ExecuteNonQuery();

                using (var cmd = new MySqlCommand("select CreatedAt from Hits order by id desc limit 10", conn))
                {
                    using (var reader = cmd.ExecuteReader())
                    {
                        var colIdx = reader.GetOrdinal("CreatedAt");
                        while (reader.Read())
                        {
                            users.Add(reader.GetString(colIdx));
                        }
                    }
                }
            }
            return users;
        }
       
    }*/
}
