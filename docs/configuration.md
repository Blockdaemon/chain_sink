# Chain sink configuration reference

## BD_CONFIG_FILES environment variable
The `BD_CONFIG_FILES` environment variable is used to set the configuration file. The value of the variable is a space-separated list of configuration files. The first file in the list is the main configuration file. The remaining files are merged into the main configuration file.

Example: one file
```bash
BD_CONFIG_FILES="config.yaml" ./chain_sink
```

Example: multiple files
```bash
BD_CONFIG_FILES="config.yaml config2.yaml config3.yaml" ./chain_sink
```

## Configuration file format
The configuration file is inferred based on the file extension. Currently the following file extensions are supported:
* `.yaml`
* `.yml`
* `.json`
* `.toml`
* `.env`

## Configuration reference

| Configuration option | Description | Type |
|-----------------------|-------------|---------------|
| `logger` | Logger configuration | `logger.Config` |
| `stream` | Stream configuration | `stream.Config` |
| `stream_count` | Number of streams to run | `integer` |
| `adapter` | Adapter configuration | `adapter.Config` |
| `metrics` | Metrics configuration | `metrics.Config` |

### `logger.Config`
Logger configuration is used to configure the logger. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `level` | Logger level | `string` | `info` |
| `disable_stacktrace` | Disable stacktrace | `boolean` | `false` |
| `development` | Development mode | `boolean` | `false` |

### `stream.Config`
Stream configuration is used to configure the stream that will be used to consume data from Chain Watch. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `url` | Stream URL | `string` | `""`  |
| `target_id` | Target ID | `string` | `""` |
| `mode` | Stream mode | `string` | `ack` |
| `headers` | Headers | `[]stream.Header` | `[]` |
| `worker_pool_size` | Worker pool size | `integer` | `1` |
| `api_key` | API key | `string` | `""` |

### `adapter.Config`
Adapter configuration is used to configure the adapter that will be used to forward the data to the target system. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `type` | Adapter type | `string` | `stdout` |
| `kafka` | Kafka configuration | `kafka.Config` | `nil` |

### `kafka.Config`
Kafka configuration is used to configure the Kafka adapter. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `producer` | Producer configuration | `kafka.ProducerConfig` | `nil` |
| `authentication` | Authentication configuration | `kafka.Authentication` | `nil` |
| `create_topic` | Create topic | `boolean` | `true` |
| `topic_name` | Topic name | `string` | `chain_sink` |
| `num_partitions` | Number of partitions | `integer` | `1` |
| `replication_factor` | Replication factor | `integer` | `1` |
| `admin_host` | Admin host | `string` | `localhost:9092` |
| `compression_type` | Compression type | `string` | `uncompressed` |
| `retention_time` | Retention time | `string` | `-1` |

### `kafka.ProducerConfig`
Producer configuration is used to configure the producer that will be used to produce the data to the Kafka topic. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `brokers` | Brokers | `[]string` | `[]` |
| `topic` | Topic | `string` | `chain_sink` |

### `kafka.Authentication`
Authentication configuration is used to configure the authentication that will be used to authenticate the producer to the Kafka cluster. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `type` | Authentication type | `string` | `none` |
| `username` | Username | `string` | `""` |
| `password` | Password | `string` | `""` |

### `metrics.Config`
Metrics configuration is used to configure the metrics server to expose prometheus metrics. The following configuration options are available:
| Configuration option | Description | Type | Default value |
|-----------------------|-------------|---------------|---------------|
| `enabled` | Enable metrics | `boolean` | `false` |
| `port` | Metrics port | `integer` | `8421` |