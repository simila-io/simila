## REST API v1

## /v1/formats
Format object:
```
{
    name: "name of the format"
}
```

### Create format
*POST /v1/formats*

```bash
curl -i -XPOST -H "content-type: application/json" -d '{"name": "pdf"}' "http://localhost:8080/v1/formats"
```
### Retrieve format
*GET /v1/formats/{name}*

```bash
curl -i -XGET "http://localhost:8080/v1/formats/pdf"
```

### Delete format
*DELETE /v1/formats/{name}*

```bash
curl -i -XDELETE "http://localhost:8080/v1/formats/pdf"
```
### List formats
*GET /v1/formats*

```bash
curl -i -XGET "http://localhost:8080/v1/formats"
```

## /v1/indexes
Index create request object:
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

### Create index 
*POST /v1/indexes*

An index may be created via providing the whole data in the `content-type: application/json` body:

```bash
curl -i -XPOST -H "content-type: application/json" -d '{"id": "1234", "format": "pdf", "tags":{"k1":"v1"}, "records": [{"id":"r1", "segment": "my text", "vector": [1, 2]}]}' "http://localhost:8080/v1/indexes"
```

or 'multipart/form-data' is also supported:

```bash
curl -i -X POST -H "content-type: multipart/form-data" -F"file=@/tmp/test.txt" -F "meta={\"id\": \"test.txt\", \"tags\":{\"k1\":\"v1\"}, \"format\": \"txt\"};type=application/json" "http://localhost:8080/v1/indexes"
```

### Update index
*PUT /v1/indexes/{id}*

```bash
curl -i -XPUT -H "content-type: application/json" -d '{"tags":{"k1":"v1"}}' "http://localhost:8080/v1/indexes/1234"
```

### Retrieve index
*GET /v1/indexes/{id}*

```bash
curl -i -XGET "http://localhost:8080/v1/indexes/1234"
```

### Delete index
*DELETE /v1/indexes/{id}*

```bash
curl -i -XDELETE "http://localhost:8080/v1/indexes/1234"
```

### Query indexes 
*GET /v1/indexes*

Query parameters:
* format={format name}
* tag={url encoded json map}
* created-after={date in "2006-01-02T15:04:05-07:00" format}
* created-before={date in "2006-01-02T15:04:05-07:00" format}
* start-index-id={starting index id, see the result object "nextPageId" field}
* limit={items per page}

```bash
curl -i -XGET "http://localhost:8080/v1/indexes?format=pdf&tags=%7B%22k1%22%3A%22v1%22%7D&created-after=2006-01-02T15:04:05-07:00&created-before=2024-01-02T15:04:05-07:00&start-index-id="123"&limit=1"
```

Query result object:
```
{
    "indexes":[],
    "nextIndexId":"",
    "total":1
}
```

### Update index records
*PATCH /v1/indexes/{id}/records*

Index records update request object:
```
{
    "upsertRecords": [{id: "record id", segment: "this is searcheable piece of the text", vector: [1, "abc", 3]}],
    "deleteRecords": [{id: "record id"}]
}
```

```bash
curl -i -XPATCH -H "content-type: application/json" -d '{"upsertRecords": [{"id": "000145f6", "segment": "this is searcheable piece of the text", "vector": [1, "abc", 3]}], "deleteRecords": [{"id": "0001044f"}]}' "http://localhost:8080/v1/indexes/test.txt/records"
```

Query result object:
```
{
    "upserted": 1,
    "deleted": 1
}
```

### Query index records
*GET /v1/indexes/{id}/records*

Query parameters:
* start-record-id={starting record id, see the result object "nextRecordId"}
* limit={items per page}

```bash
curl -i -XGET "http://localhost:8080/v1/indexes/test.txt/records?start-record-id=eyJpbmRleF9pZCI6InRlc3QudHh0IiwicmVjb3JkX2lkIjoiMDAwMDAwMDIifQ==&limit=1"
```

Query result object:
```
{
    "records": [],
    "nextRecordId": "",
    "total": 1
}
```

## /v1/search
Search request object:
```
{
    "text": "a list of words that should be found",
    "tags": "a json map of {"key":"val"} tags to filter by",
    "indexIDs": "a list of index IDs to filter by",
    "distinct": "if true, the result will contain only one record(first) per index",
    "pageId": "starting page id, see the result object "nextPageId" field",
    "limit": "items per page"
}
```

### Search records
*POST /v1/search*

```bash
curl -i -XPOST -H "content-type: application/json" -d '{"text": "shakespeare", "tags":{"k1":"v1"}, "indexIDs":["test.txt"], "pageId":"eyJpbmRleF9pZCI6InRlc3QudHh0IiwicmVjb3JkX2lkIjoiMDAwMGJhODUifQ=="}' "http://localhost:8080/v1/search"
```

Search result object:
```
{
    "records": [],
    "nextPageId": "",
    "total": 1
}
```