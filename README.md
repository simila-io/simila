[![build](https://github.com/simila-io/simila/actions/workflows/build.yaml/badge.svg)](https://github.com/simila-io/simila/actions/workflows/build.yaml) [![docker](https://github.com/simila-io/simila/actions/workflows/docker.yaml/badge.svg)](https://github.com/simila-io/simila/actions/workflows/docker.yaml) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/simila-io/simila/blob/master/LICENSE)

## Simila 
Simila enables syntactic and semantic search across custom data sources.

## Features
- Syntactic full-text search capabilities leveraged by popular search engines.
- Semantic search powered by large language models.
- Flexible deployment configurations — run on-premises or use cloud services.
- Unified API for seamless integration.
- Tools for search through structured and unstructured data sources.
- Ability to search through both private and publicly available data sources.

## General information
Simila is a service written in Golang. It exposes RESTful and gRPC-based APIs to facilitate data ingestion and text search capabilities. Under the hood, Simila provides syntactic search functionality, leveraging popular full-text search engines such as Postgres (supported), Elasticsearch, Bleve, etc. For semantic search, a Large Language Model (LLM) may be employed to generate embeddings, enabling semantic search over a vector database (currently supported on Postgres).

To run Simila, you need at least two components: the Simila service and the Simila database. The Simila service is an executable that can be built from the source in the repository or downloaded as a Docker container. The Simila database is a component that supports both full-text and vector search. In a minimalistic configuration, Postgres can be used as the Simila database.

Semantic search capabilities are provided by integrating with an LLM model for calculating embeddings. These embeddings are then stored in the Simila database and being used for the semantic search.

## Core concepts
The Simila design and domain objects are described in the [core concepts](docs/concepts.md) section. It is recommended to read this section before starting to use Simila.

## API
The Simila API is available over [gRPC](api/proto) and [HTTP](api/openapi/README.md) protocols.

Calling APIs example:

```bash
curl localhost:8080/v1/ping # http API
grpcurl --plaintext localhost:50051 grpc.health.v1.Health/Check # grpc API 
```

## Configuration
The [configuration settings](docs/configuration.md) could be passed to Simila via configuration file `simila start --config simila.yaml` or via the environment variables starting with the `SIMILA_` prefix.

## How to run
Here are some examples of how to run Simila:

- [Runing localy in docker compose](docs/deployment.md#docker-compose-locally)
- [Build and run locally from the source code](docs/deployment.md#compile-from-the-source-code-and-run-it-locally)

## CLI
The Simila service has its own CLI client called `scli`.

Install and run the CLI:

```bash
curl -s https://raw.githubusercontent.com/simila-io/simila/main/install-cli | bash -s -- -d /tmp # install `scli` client
/tmp/scli --addr localhost:50051 # connect to simila service
```

For more details please check [this](docs/cli.md) section.

## License
Apache License 2.0, see [LICENSE](LICENSE).
