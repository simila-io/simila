![Build](https://github.com/simila-io/simila/actions/workflows/build.yaml/badge.svg) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/simila-io/simila/blob/master/LICENSE)
# simila
Simila search service

## Quick start

### 1. Run DB docker

```bash
make db-start
```
### 2. Create Simila DB

```bash
PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres -d postgres -c "create database simila"
```

### 3. Run Simila executable

```bash
make run
```

## Notes

- Simila uses Postgres with the PGroonga extention enabled https://hub.docker.com/r/groonga/pgroonga. More info on PGroonga: https://pgroonga.github.io/reference/.
