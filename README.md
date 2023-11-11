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

## API
The Simila API is based on the [concepts](docs/concepts.md) and available over [gRPC](api/proto) and [HTTP](api/openapi/README.md) protocols.

## How to run
Here are some examples, how Simila can be run:

- [Runing localy in docker compose](docs/deployment.md#docker-compose-locally)
- [Build and run locally from the source code](docs/deployment.md#compile-from-the-source-code-and-run-it-locally)

## License
Apache License 2.0, see [LICENSE](LICENSE).

