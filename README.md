![Build](https://github.com/simila-io/simila/actions/workflows/build.yaml/badge.svg) ![Docker](https://github.com/simila-io/simila/actions/workflows/docker.yaml/badge.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/simila-io/simila/blob/master/LICENSE)
# simila
Simila search service

## Quick start

### 1. Start Simila

```bash
make compose-up
```

### 2. Tail logs (optional)

```bash
make compose-logs
```

### 3. Use the API
```bash
curl localhost:8081/v1/ping # http API
grpcurl --plaintext localhost:50052 grpc.health.v1.Health/Check # grpc API 
```

Check the [API documentation](pkg/api/README.md) for details.

### 4. Stop Simila

```bash
make compose-down
```
