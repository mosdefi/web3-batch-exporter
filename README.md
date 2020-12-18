# web3-batch-exporter

This is a **highly experimental** (:see_no_evil:) service that reads current and historical state variables about any smart contract on the Ethererum blockchain and pushes them to a timeseries database. Currently it's using [victoria-metrics](https://github.com/VictoriaMetrics/VictoriaMetrics) as its datasink as this supports both live and historical data loads.

It integrates with the following services:
- [web3-batch-service](https://github.com/mosdefi/web3-batch-service) // node bridge to [web3-batch-call](https://github.com/x48-crypto/web3-batch-call)
- [infura](https://infura.io) or [alchemy](https://alchemyapi.io/)    // JSON-RPC endpoint provider to Ethererum
- [etherscan](https://etherscan.io)                                   // used to obtain ABIs

The service is exposing the following JSON endpoints on port `8000`:
- `/live` for current state changes this creates a Prometheus exporter which can be scraped by prometheus or victoria-metrics. This runs in an endless loop until you stop it.
- `/metrics` For compatibility reasons, there is also a prometheus exporter that exposes live data which can be used by a prometheus scraper.
- `/historical` for historical data you have to provide the query params `startBlock` and `endBlock` so that all blocks between will be scraped.

## Configuration
To demonstrate its usage, there is `stack` dir with a `docker-compose.yml` file which contains all the cookies.

First it's necessary to provide the two ENV variables:
- `PROVIDER_URL`
- `ETHERSCAN_API_KEY`

The easiest way is to add them to your local `stack/.env` file.

## Build & Run
`cd stack && ./run.sh`

## Examples
To start the request with all [yearn](https://yearn.finance) vaults and strategies there is a `stack/sample_request.json` which can be tested with:

### Live data
`curl -X POST -H 'Content-Type: application/json' -d@stack/sample_request.json "http://localhost:8000/live"`
Once started, there will be a new batched call and a push to victoria-metrics with the results every minute.

### Historical data
`curl -X POST -H 'Content-Type: application/json' -d@stack/sample_request.json "http://localhost:8000/historical?startBlock=11460686&endBlock=11460690"`
This will query all the the requested blocks and push them to the timeseries db.


## Dockerized setup
The following containers are started:

- `web3-batch-exporter` // the main container with the main service from this repo
- `web3-batch-service`  // node-service bridge to communicate with the blockchain
- `victoria-metrics`    // the timeseries db
- `vmagent`             // the CLI tool that scrapes the metrics when using live mode
- `grafana`             // visualize your data

The data can be viewed directly in Grafana under: [http://localhost:3000](http://localhost:3000) .

Here's a small screenshot:
![grafana_dashboard.png](https://raw.githubusercontent.com/mosdefi/web3-batch-exporter/main/grafana/grafana_dashboard.png)


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
- write some tests :joy: 
