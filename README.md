# web3-batch-exporter

This is a **highly experimental** (:see_no_evil:) service that continuously tracks state about any [yEarn](https://yearn.finance) contract and pushes it to [prometheus](https://prometheus.io).

It integrates with the following services:
- [web3-batch-service](https://github.com/mosdefi/web3-batch-service)
- [infura](https://infura.io)
- [etherscan](https://etherscan.io)

By default, the service is exposing a JSON endpoint on port `8000` which can be used as a proxy and mini-job system for running web3-batch-call requests.

## Configuration
To demonstrate its usage, there is `stack` dir with a `docker-compose.yml` file which contains all the cookies.

First it's necessary to provide the two ENV variables:
- `PROVIDER_URL`
- `ETHERSCAN_API_KEY`

The easiest way is to add them to your local `stack/.env` file.

## Build & Run
`cd stack && ./run.sh`

## POSTing data to prometheus
To start the request with all yearn vaults and strategies there is a `stack/sample_request.json` which can be tested with:

`cd stack && ./post.sh`

Once started, there will be a new batched call and a push to prometheus with the results every minute.

You can CTRL-C this safely, the worker will continue querying and pushing data in the background.

```
calling web3-batch-service...
2020/12/08 19:40:25 Successfully exported 206 metrics to prometheus push gateway.
```

## Viewing data
The dockerized service is configured to push the data to prometheus push gateway which is scraped by prometheus.
The data can be viewed directly in prometheus under: [http://localhost:9090](http://localhost:9090) or in Grafana under: [http://localhost:3000](http://localhost:3000) .
The datapoints are labeled starting with their namespace like the following example:

`strategies_0x25fAcA21dd2Ad7eDB3a027d543e617496820d8d6_balanceOf` references the `balanceOf` field of the `StrategyVaultUSDC`

There is a `grafana_dashboard.json` file that can be loaded into the JSON model of grafana containing the following metrics:

![grafana_dashboard.png](https://raw.githubusercontent.com/mosdefi/web3-batch-exporter/main/grafana/grafana_dashboard.png)


## Control flow
The control flow is to POST a JSON to the service which registers a worker that runs the posted queries every minute until another request arrives.
The worker itself first queries the web3-batch calls to the Ethereum blockchain, then parses and pushes any float64-convertible values to prometheus.


## Auto-conversions
Currently the following conversions are taking place so that the values can be posted as prometheus gauges as float values.
- bigint quoted as string is converted to float64
- int is converted to float64
- bool is converted to 1.0 or 0.0 float64

There is also some experimental scaling for some numerical values if the contract provides a `decimals` field.
If so, the number is divided by `10^decimals`. This needs to be improved for strategies where there is no `decimals` field but instead the corresponding vault has this field.
Maybe this could be done by some external lookup table or so.

## Limitations
There are some `TODO`s in the code namely:
- parse array results not only the first result
- parse nested results
- maintain the scaling factor
- add human-readable aliases to the gauges instead of the contract addresses
