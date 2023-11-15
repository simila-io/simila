## How to run
The section describes how to run Simila quickly.

### Docker-compose locally

Use `make` to build the artifacts and run everything in the Docker:
```bash
make compose-up
```

Tail logs (optional):
```bash
make compose-logs
```

Test API is up and running (optional):
```bash
curl localhost:8080/v1/ping
```

To stop Simila run the command:
```bash
make compose-down
```

### Compile from the source code and run it locally
**NOTE**: you need `Golang`, `docker` and `make` to be installed  

Compile the `simila` and `scli` executables (they will be put into `build/` directory):
```bash
make build
```

Run postgres in the docker container:
```bash
make db-start
```

Start the Simila service:
```bash
./build/simila start
```

Connect to Simila service using `scli` command tool:
```bash
./build/scli 
```

To stop postgres:
```bash
make db-stop
```
