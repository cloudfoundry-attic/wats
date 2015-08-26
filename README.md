# Windows Acceptance Tests (WATs)
![Angkor Wat](http://upload.wikimedia.org/wikipedia/commons/thumb/f/f5/Buddhist_monks_in_front_of_the_Angkor_Wat.jpg/640px-Buddhist_monks_in_front_of_the_Angkor_Wat.jpg)

## Running the tests

### Test Setup

To run the Diego Acceptance tests, you will need:
- a running CF instance
- credentials for an Admin user
- an environment variable `$CONFIG` which points to a `.json` file that contains the application domain

You can use `scripts/bosh_lite_config.json` to run the specs against
[bosh-lite](https://github.com/cloudfoundry/bosh-lite). Replace
credentials and URLs as appropriate for your environment.

**NOTE**: The secure_address must be some inaccessible endpoint from
  any container, e.g., an etcd endpoint

### Running the tests against a bosh-lite

`./scripts/bosh_lite_run_wats.sh`

### Self signed certificates

If you are running the tests with version newer than `6.0.2-0bba99f`
of the Go CLI against bosh-lite or any other environment using
self-signed certificates, add the following

```
  "skip_ssl_validation": true
```

to your config file as well.

### References

- [cats]: https://github.com/cloudfoundry/cf-acceptance-tests
