# Windows Acceptance Tests (WATs)
![Angkor Wat](http://upload.wikimedia.org/wikipedia/commons/thumb/f/f5/Buddhist_monks_in_front_of_the_Angkor_Wat.jpg/640px-Buddhist_monks_in_front_of_the_Angkor_Wat.jpg)

## Running the tests

### Test Setup

To run the Diego Acceptance tests, you will need:
- a running CF instance
- credentials for an Admin user
- an environment variable `$CONFIG` which points to a `.json` file that contains the application domain

The following script will configure these prerequisites for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
installation. Replace credentials and URLs as appropriate for your environment.

NOTE: The secure_address must be some inaccessible endpoint from any container, e.g., an etcd endpoint

```bash
#! /bin/bash

cat > integration_config.json <<EOF
{
  "api": "api.10.244.0.34.xip.io",
  "admin_user": "admin",
  "admin_password": "admin",
  "apps_domain": "10.244.0.34.xip.io",
  "secure_address": "10.244.16.2:4001",
  "skip_ssl_validation": true
}
EOF
export CONFIG=$PWD/integration_config.json
```

If you are running the tests with version newer than 6.0.2-0bba99f of the Go CLI against bosh-lite or any other environment
using self-signed certificates, add

```
  "skip_ssl_validation": true
```

to your integration_config.json as well.

[cats]: https://github.com/cloudfoundry/cf-acceptance-tests
