![Build](https://github.com/simila-io/simila/actions/workflows/build.yaml/badge.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/simila-io/simila/blob/master/LICENSE)
# simila
Simila search service

## Quick start

### 1. Run Simila

```bash
make compose-up
```

### 2. Use the API
```bash
curl localhost:8081/v1/ping # http API
grpcurl --plaintext localhost:50052 grpc.health.v1.Health/Check # grpc API 
```

### 3. Tail logs (optional)

```bash
make compose-logs
```

### 4. Stop Simila

```bash
make compose-down
```

## Notes

- Simila uses Postgres with the PGroonga extention enabled https://hub.docker.com/r/groonga/pgroonga.  
  More info on PGroonga: https://pgroonga.github.io/reference/ and Groonga engine https://groonga.org/docs/reference.
