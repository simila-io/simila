# Configuration
The section describes Simila configuration settings.

## Settings

### GrpcTransport
This group of settings specifies GRPC API parameters. If `Address` is set to an empty string the API is listened on the `Port` for all the network interfaces.

### HttpPort
This parameter specifies on which port the HTTP API is listened

### SearchEngine
This parameter defines which search engine, behind Simila API, is used for indexing and search. At the moment Simila supports Postgres >= v15 only in 3 different modes: `pgroonga`, `pgtrigram` and `pgfts`.

In order to use `pgroonga` and `pgtrigram` modes the `pgroonga` and `pg_tgrm` Postgres extensions must be installed accordingly.
Depending on the chosen mode the query syntax used in search API is different:

- `pgroonga`: see https://pgroonga.github.io/reference/operators/query-v2.html
- `pgfts`: see https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-PARSING-QUERIES for `websearch_to_tsquery()`
- `pgtrigram`: no query language, the query text is matched against the stored index using the `trigram word similarity` criteria,


### DB
This group of settings specifies the Simila DB settings. At the moment only Postgres >= v15 is supported. The `SSLMode` param can be set to `disable` or `require` depending on the environment and desired SSL mode (localhost, RDS, etc.)

## Examples

### Configuration file

```json
{
  "GrpcTransport": {
    "Network": "tcp",
    "Address": "",
    "Port": 50051
  },
  "HttpPort": 8080,
  "SearchEngine": "pgfts",
  "DB": {
    "Driver": "postgres",
    "Host": "localhost",
    "Port": "5432",
    "Username": "postgres",
    "Password": "postgres",
    "DBName": "simila",
    "SSLMode": "disable"
  }
}
```

### Environment variables

```bash
SIMILA_DB_DRIVER=postgres
SIMILA_DB_HOST=127.0.0.1
SIMILA_DB_PORT=5432
SIMILA_DB_USERNAME=postgres
SIMILA_DB_PASSWORD=postgres
SIMILA_DB_DBNAME=simila
SIMILA_DB_SSLMODE=disable 
SIMILA_SEARCHENGINE=pgroonga
SIMILA_GRPCTRANSPORT_PORT=50051
SIMILA_HTTPPORT=8080
```
