# Windows Acceptance Tests (WATs)
![Angkor Wat](http://upload.wikimedia.org/wikipedia/commons/thumb/f/f5/Buddhist_monks_in_front_of_the_Angkor_Wat.jpg/640px-Buddhist_monks_in_front_of_the_Angkor_Wat.jpg)

## Running the tests

### Test Setup

To run the Windows Acceptance tests, you will need:
- a running CF instance
- credentials for an Admin user
- an environment variable `CONFIG` which points to a `.json` file that contains the application domain
  - an example configuration can be found [here](scripts/integration_config.json)

**NOTE**: The secure_address must be some inaccessible endpoint from
  any container, e.g., a BOSH director endpoint

### Running the tests

`./scripts/run_wats.sh ./scripts/integration_config.json`

or

`CONFIG=<path-to-config.json> ginkgo -r`

### Self signed certificates

If you are running the tests with version newer than `6.0.2-0bba99f`
of the Go CLI against an environment using self-signed certificates,
add the following:

```
  "skip_ssl_validation": true
```

to your config file as well.

