# Chain sink quick start

This guide will walk you through the process of setting up chain sink and consuming data from Chain Watch. This guide assumes basic understanding of Blockdaemon Chain Watch. If you are not familiar with Chain Watch, please see the [Chain Watch documentation](https://docs.blockdaemon.com/docs/overview-events) for more information.

## Obtaining chain sink

You can download the latest release of chain sink from the [releases page](https://github.com/Blockdaemon/chain_sink/releases).

Alternatively you can build chain sink from source by running the following command:
```bash
make build
```

This will build the chain sink binary and place it in the `out` directory. 

## Creating a websocket target
Chain sink requires a websocket target to be created in Chain Watch. You can create a websocket target by using the [Chain Watch targets API](https://docs.blockdaemon.com/reference/add-target).
```bash
curl -X POST "https://svc.blockdaemon.com/streaming/v2/targets" \
  -H "x-api-key: <your-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
        "description": "Chain sink websocket quick start",
        "max_buffer_count": 100,
        "name": "Chain sink websocket quick start",
        "settings": {
            "noack":false
        },
        "type": "websocket"  
      }'
```

The response will contain the target ID. You will need to use this target ID in the chain sink configuration.
```json
{
    "created_at": "2026-01-28T11:01:21.931253Z",
    "current_buffer_count": 0,
    "description": "Chain sink websocket quick start",
    "id": "e89d7256-174f-429d-9a79-13ef57fb9e4f",
    "max_buffer_count": 100,
    "name": "Chain sink websocket quick start",
    "settings": {
        "noack": false
    },
    "status": "active",
    "status_since": "2026-01-28T11:01:21.931253Z",
    "type": "websocket",
    "updated_at": "2026-01-28T11:01:21.931253Z"
}
```
## Understanding acknowledgement mode
Chain watch supports two modes of operation: acknowledgement mode and no acknowledgement mode.
* `settings.noack=true`: no acknowledgement mode. In this mode the client will not have to send any acknowledgement messages to the server. The server will send the next message immediately after the previous message is received. This also means that if the client fails to process the message, the message will be lost. This mode will achieve the highest throughput, but is not recommended when data integrity is important.
* `settings.noack=false`: acknowledgement mode. In this mode the client has to send an acknowledgement message for each message received from the server. The server will only send the next message after the acknowledgement message is received. This mode is more reliable, but slower than no acknowledgement mode.

Chain sink supports both modes of operation, the default is acknowledgement mode. When configuring chain sink it is important that the target mode set in chain sink matches the target mode set in Chain Watch.

## Configuring chain sink
This guide assumes a rule has been created in Chain Watch using the target that has been created in the previous step. This rule will produce events that can be consumed by chain sink using the websocket target.

Below you find an example configuration for chain sink for this particular example in yaml:
```yaml
stream:
  url: "wss://svc.blockdaemon.com/streaming/v2/targets/e89d7256-174f-429d-9a79-13ef57fb9e4f/websocket"
  mode: "ack"
  api_key: "<your-api-key>"

adapter:
    type: "stdout"
```

Important notes:
* The `url` is the websocket URL for the target. The url is in the format `wss://svc.blockdaemon.com/streaming/v2/targets/<target_id>/websocket`.
* The `api_key` is the API key for the Chain Watch API. You can get the API key from the Blockdaemon webapp.
* The `mode` is the mode of operation for the target. This mode should match the target mode set in Chain Watch.
* The `adapter` is the adapter that will be used to forward the data to the target system. The adapter is set in the adapter configuration. For this quick guide we use the stdout adapter, which will print the data to the console. This is not recommended for production use, but is useful for testing.

## Running chain sink
To run chain sink, use the following command:
```bash
BD_CONFIG_FILES="config.yaml" ./out/chain_sink
```
