![Blockdaemon logo](https://start.blockdaemon.com/hubfs/Branding/Blockdaemon%20Email%20Header.png)


# Chain Sink

Chain Sink is a high performance data sink using Blockdaemon Chain Watch websocket API. 

## Chain Watch API
For more information about the Chain Watch API, please see:
* [Quick start](https://docs.blockdaemon.com/docs/overview-events)
* [API reference](https://docs.blockdaemon.com/reference/get-event-protocol-overview)

## How Chain Sink works
Chain Sink is a single binary that can be used to consume data from Chain Watch using websockets, and forward the data to one of the supported adapters. Currently the following adapters are supported:
* <b>stdout</b>: prints the data to the console
* <b>kafka</b>: produces the data to a Kafka topic

## Configuration
Chain Sink is fully configuration driven. Please see the [configuration reference](./docs/configuration.md) for more information.

## Getting started
See our [quick start guide](./docs/quick-start.md) to get started or see our [examples](./docs/examples) for more information.