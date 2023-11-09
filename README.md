[![build](https://github.com/simila-io/simila/actions/workflows/build.yaml/badge.svg)](https://github.com/simila-io/simila/actions/workflows/build.yaml) [![docker](https://github.com/simila-io/simila/actions/workflows/docker.yaml/badge.svg)](https://github.com/simila-io/simila/actions/workflows/docker.yaml) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/simila-io/simila/blob/master/LICENSE)

# Simila 
Simila enables syntactic and semantic search across custom data sources.

# Features
- Syntactic full-text search capabilities leveraged by popular search engines
- Semantic search leveraged by using large language models
- Flexible deployment configuration - running on premises or use cloud services
- Unified API
- Tools for scanning structured and unstructured data sources
- To search through the private and publicly available data sources

## General information
Simila is a service written in Golang. It exposes RESTful and gRPC-based APIs to facilitate data ingestion and text search capabilities. Under the hood, Simila provides syntactic search functionality, leveraging popular full-text search engines such as Postgres (supported), Elasticsearch, Bleve, etc. For semantic search, a Large Language Model (LLM) may be employed to generate embeddings, enabling semantic search over a vector database (currently supported on Postgres).

To run Simila you need at least two components - Simila service and the Simila database. The Simila service is an executable which can be built from the source of the repository or been downloaded as a docker container. The Simila database is the component, which supports full-text and the vector search. In minimalistic configuration Postgres can be used as Simila database.

The semantic search capabilities provided by using an integration with an LLM model for calculations embeddings. The embeddings are stored in the Simila database.

## Deployment configurations
This section briefly describes possible deployment configurations.

### Local dev environment
To run Simila on the local machine you may use the supplied docker-compose configuration or run the Simila service and Posgres locally.

### Self-hosted
The self-hosted environment is similar to the local dev environment, and it requires both components Simila service and Postres managed by your favorite tool (k8s, terraform etc.) or run in your premises manually.

### AWS RDS
In the AWS cloud Simila service may be run as an instance or as a part of k8s cluster container. The Simila database may be run as RDS service (Postgres)

## Quick start

### Use the docker-compose to start it locally

Use `make` to build the artifacts and run everything in the Docker:
```bash
make compose-up
```

Tail logs (optional):
```bash
make compose-logs
```

Use the API:
```bash
curl localhost:8081/v1/ping # http API
grpcurl --plaintext localhost:50052 grpc.health.v1.Health/Check # grpc API 
```

To Stop Simila run the command:
```bash
make compose-down
```


### Compile from the source code and run it locally
NOTE: you need Golang, docker and make be installed

Compile the Simila and scli executables (they will be put into `build/` directory) 
```bash
make build
```

Run postgres in the docker container
```bash
make db-start
```

Start the simila service:
```bash
./build/simila start
```

Connect the simila service using `scli` command tool:
```bash
./build/scli 
```
