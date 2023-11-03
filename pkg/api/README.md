## REST API v1

## /v1/formats
Format object:
```
{
    name: "pdf"
}
```

### Create a format

```bash
curl -XPOST -H "content-type: application/json" -d '{"name": "pdf"}' http://localhost:8080/v1/formats
```
### Retrieve a format

```bash
curl -XGET http://localhost:8080/v1/formats/{name}
```

### Delete a format

```bash
curl -XDELETE http://localhost:8080/v1/formats/{name}
```
### List all formats

```bash
curl -XGET http://localhost:8080/v1/formats
```

## /v1/indexes
Index object:
```
{
    id: "the index ID"
    format: "pdf"
    tags: ["userABC", "salesTeam"]
    document: "a base64 encoded document, used for create new index only" 
    records: [{id: "abcd", segment: "hello world", vector: [1, 2]}]
}
```

An index record object:
```
{
    id: "a base64 encoded vector"
    segment: "this is searcheable piece of the text"
    vector: [1, "abc", 3]
}
```

### Create new index 
POST /v1/indexes

An index may be created via providing the whole data in the `content-type: application/json` body:
```bash
curl -XPOST -H "content-type: application/json" -d '{"id": "1234", "format": "pdf"}' http://localhost:8080/v1/indexes
```

or 'multipart/form-data' is also supported:

example:
```bash
curl -i -X POST -H "content-type: multipart/form-data" -F"file=@/Users/user/Downloads/fr_9782_size1024.jpg" -F "meta={\"id\": \"1234\", \"format\": \"jpg\"};type=application/json" http://localhost:8080/v1/indexes
```
