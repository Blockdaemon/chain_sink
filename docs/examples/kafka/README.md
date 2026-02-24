# Chain sink with kafka

This example shows how to configure chain sink to forward data to a Kafka topic. This guide assumes basic understanding of Kafka, Chain Watch and chain sink For more information see:
* [Kafka documentation](https://kafka.apache.org/documentation/)
* [Chain Watch documentation](https://docs.blockdaemon.com/docs/overview-events)
* [chain sink quick start](../../quick-start.md)

## Requirements
* A websocket target created in Chain Watch. Please see the [chain sink quick start](../../quick-start.md) for more information.
* A rule that produces events to the websocket target
* Docker with docker compose installed, or an already running Kafka cluster
* The `chain_sink` binary

## Kafka in docker
This example uses Kafka ran in docker. In production you should use a Kafka cluster, which is either self hosted or managed by a cloud provider.
To follow along with this example, you can start a Kafka cluster in docker by running the following command:
```bash
docker compose -f docker-compose.kafka.yaml up -d
```

This will start a Kafka cluster with a schema registry and a control center.

## Chain sink configuration
The example configuration is located in the `example.config.yaml` file. 
* Make sure to replace the `<target_id>` and `<api_key>` with the actual values.
* If you do not have a target created yet, please see the [chain sink quick start](../../quick-start.md) for more information.
* The example configuration is configured to forward data to the `chain_sink_example` topic. You can change the topic name to your own topic name if you want. Chain sink will create the topic if it does not exist, this can be disabled by setting the `create_topic` flag to `false`.
* You can change the brokers to your own Kafka brokers if you want. See the [configuration reference](../../configuration.md) for more information about the configuration options for the Kafka adapter.

## Running chain sink
To run chain sink, use the following command:
```bash
BD_CONFIG_FILES="example.config.yaml" ./out/chain_sink
```

## Verifying the data in Kafka
The local version of Kafka is running in docker, and provides a control center UI which can be accessed at http://localhost:9021.

You can verify that the data is being forwarded to Kafka by checking the `chain_sink_example` topic in the control center.

## Cleaning up
To clean up the docker resources, you can run the following command:
```bash
docker compose -f docker-compose.kafka.yaml down
```