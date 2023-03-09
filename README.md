# :telescope: Lookout

> Keeps a lookout for lookup quality

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
$ recorder --help
Usage of recorder:
  -checkInterval duration
        The interval at which checks are run. (default 10m0s)
  -logLevel string
        The logging level. Only applied if GOLOG_LOG_LEVEL environment variable is unset. (default "info")
```

## License

[SPDX-License-Identifier: Apache-2.0 OR MIT](LICENSE.md)
