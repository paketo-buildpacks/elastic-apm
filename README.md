# Paketo Buildpack for Elastic APM

## Buildpack ID: `paketo-buildpacks/elastic-apm`
## Registry URLs: `docker.io/paketobuildpacks/elastic-apm`
The Paketo Buildpack for Elastic APM is a Cloud Native Buildpack that contributes the [Elastic APM][e] Agent and configures it to connect to the service.

[e]: https://www.elastic.co/solutions/apm

## Behavior
This buildpack will participate if all the following conditions are met

* A binding exists with `type` of `ElasticAPM`

The buildpack will do the following for Java applications:

* Contributes a Java agent to a layer and configures `$JAVA_TOOL_OPTIONS` to use it
* Transforms the contents of the binding secret to environment variables with the pattern `ELASTIC_APM_<KEY>=<VALUE>`

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`
| Key                   | Value   | Description                                                                                       |
| --------------------- | ------- | ------------------------------------------------------------------------------------------------- |
| `<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>` |

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0
