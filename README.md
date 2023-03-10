# :telescope: Lookout

[![Go Test](https://github.com/ipni/lookout/actions/workflows/go-test.yml/badge.svg)](https://github.com/ipni/lookout/actions/workflows/go-test.yml)

> On the lookout for lookup quality

This repo implements a service that continuously performs a set of lookup checks using a sample set
on one or more IPNI endpoints.
It is built to be highly extensible, offering the ability to programmatically define checks as well
as samplers. The service then exposes the coverage and latency metrics as Prometheus metrics tagged
with checker and sampler names.

## Install

To install the event recorder run:

```shell
go install github.com/ipni/lookout/cmd/lookout@latest
```

## Usage

```text
$ lookout --help
Usage of lookout:
  -config string
        The path to lookout YAML config file. (default "config.yaml")
  -logLevel string
        The logging level. Only applied if GOLOG_LOG_LEVEL environment variable is unset. (default "info")
```

### Config

The `lookout` config must be specified as `--config` flag, with value pointing to a valid
configuration
YAML file. An example config can be found at [`examples/config.yaml`](examples/confg.yaml)

* `checkers` - Set of checkers to use.
    * `<checker-name>` - The name to associate to the checker, which will appear in metric tags with key `checker`.
        * `type` - The type of checker to use. Only `ipni-non-streaming` is currently supported.
        * `ipniEndpoint` - The HTTP URL of IPNI compatible lookup API to check.
        * `Timeout` - The timeout for each multihash lookup.
        * `ipfsDhtCascade` - Whether to request cascading over IPFS DHT
        * `parallelism` - The number of concurrent lookups to check against the endpoint.
* `samplers` - Set of samplers to use for generating multihash lookup samples
    * `<sampler-name>` - The name to associate to the sampler, which will appear in metric tags with key `sampler`.
        * `type` - The type of sampler to use. Only `saturn-orch-top-cids` and `awesome-ipfs-datasets` are supported.
* `checkInterval` - The interval at which to run checks.
* `checkersParallelism` - The maximum number of concurrent checkers to run in each cycle.
* `samplersParallelism` - The maximum number of concurrent samplers to run in each cycle.
* `metricsListenAddr` - The listen address of the metrics HTTP server.

The check cycle is then repeated at the configured interval for all permutations of the configured `checkers` and `samplers`.

An example config can be found at [`examples/config.yaml`](examples/confg.yaml)

## License

[SPDX-License-Identifier: Apache-2.0 OR MIT](LICENSE.md)
