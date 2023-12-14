## Example

This section describes an example of how to work with Simila API and objects. The example assumes a situation when a company has 2 tenant organizations `Coca-Cola Company` and `Ford Motors Company`, both organizations have a balance file `balance.xlsx` with some credit and debit records for 2023. The company wants to make the data of the tenant organizations to be searchable.

The example illustrates:

- how to make different types of data searchable;
- how to add, update and delete searchable data;
- how to perform a search across one or many nodes;
- how to limit the scope of searchable data with `tags` and `formats` filters;
- how to delete big chunks of searchable data;

If some concepts are not clear, please refer to [core concepts](../docs/concepts.md).   
If some API is not described in this section, please refer to [simila.yaml](../api/openapi/v1/simila.yaml).

### 1. Create formats for records

Create format for organization meta records (e.g. name):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"name\": \"organizationsMeta\", \"basis\": \"`echo '[{"name": "table", "type": "string"}, {"name": "id", "type": "integer"}, {"name": "column", "type": "string"}]' | base64`\"}" "http://localhost:8080/v1/formats" | jq
```

Create format for spreadsheet meta records (e.g. filename):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"name\": \"spreadsheetsMeta\", \"basis\": \"`echo '[{"name": "path", "type": "string"}, {"name": "filename", "type": "string"}]' | base64`\"}" "http://localhost:8080/v1/formats" | jq
```

Create format for spreadsheet data records (e.g. cells data):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"name\": \"spreadsheetsData\", \"basis\": \"`echo '[{"name": "sheet", "type": "string"}, {"name": "row", "type": "string"}, {"name": "col", "type": "string"}]' | base64`\"}" "http://localhost:8080/v1/formats" | jq
```

### 2. List the created formats

List the created formats for the aforementioned record types:

```bash
curl -s -XGET "http://localhost:8080/v1/formats" | jq
```

### 3. Create searchable records (and corresponding nodes)

Make `Coca-Cola Company` organization name searchable for internal use only (the `public` tag):

```bash 
curl -s -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"folder\", \"tags\": {\"public\": \"false\"}, \"records\": [{\"id\":\"organizations|1234|name\", \"format\": \"organizationsMeta\", \"segment\": \"Coca-Cola Company\", \"rankMultiplier\": 2.0, \"vector\": \"`echo '["organizations", "1234", "name"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234/records" | jq
```

Make `Coca-Cola Company` organization balance spreadsheet searchable (both meta and data) for internal use only (the `public` tag):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"document\", \"tags\": {\"public\": \"false\"}, \"records\": [{\"id\":\"/spreadsheets/2023|balance.xlsx|name\", \"format\": \"spreadsheetsMeta\", \"segment\": \"company balance 2023\", \"rankMultiplier\": 1.5, \"vector\": \"`echo '["/spreadsheets/2023", "balance.xlsx"]' | base64`\"}, {\"id\":\"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$100\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\":\"debit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$200\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R2", "C1"]' | base64`\"}, {\"id\":\"credit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$1000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R1", "C1"]' | base64`\"}, {\"id\":\"credit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$2000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R2", "C1"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234%2Fbalance.xlsx/records" | jq
```

Make `Ford Motors Company` organization name searchable for everyone (the `public` tag):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"folder\", \"tags\": {\"public\": \"true\"}, \"records\": [{\"id\":\"organizations|5678|name\", \"format\": \"organizationsMeta\", \"segment\": \"Ford Motors Company\", \"rankMultiplier\": 2.0, \"vector\": \"`echo '["organizations", "5678", "name"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F5678/records" | jq
```

Make `Ford Motors Company` organization balance spreadsheet searchable (both meta and data) for everyone (the `public` tag):

```bash
curl -s -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"document\", \"tags\": {\"public\": \"true\"}, \"records\": [{\"id\":\"/spreadsheets/2023|balance.xlsx|name\", \"format\": \"spreadsheetsMeta\", \"segment\": \"company balance 2023\", \"rankMultiplier\": 1.5, \"vector\": \"`echo '["/spreadsheets/2023", "balance.xlsx"]' | base64`\"}, {\"id\":\"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$300\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\":\"debit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$600\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R2", "C1"]' | base64`\"}, {\"id\":\"credit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$3000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R1", "C1"]' | base64`\"}, {\"id\":\"credit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$6000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R2", "C1"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F5678%2Fbalance.xlsx/records" | jq
```

### 4. List the created nodes

List organization node children:

```bash
curl -s -XGET "http://localhost:8080/v1/nodes?path=/orgs" | jq
```

List `Coca-Cola Company` node children:

```bash
curl -s -XGET "http://localhost:8080/v1/nodes?path=/orgs/1234" | jq
```

List `Ford Motors Company` node children

```bash
curl -s -XGET "http://localhost:8080/v1/nodes?path=/orgs/5678" | jq
```

### 5. Search created records by specifying different paths, tags etc.

Search the organization node children, return the most relevant record per node only:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "tags": {}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```

Search the organization node children with filter via tags (the `public` tag), return the most relevant record per node only:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "tags": {"public":"true"}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```

Search the organization node children with filter via format (record type), return the most relevant record per node only:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "format": "organizationsMeta", "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```

Search the `Coca-Cola Company` balance node only, return all the matched records for the node:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234/balance.xlsx", "tags": {}, "strict":true, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```

### 6. Add, update and delete searchable records of a node

Modify the `Coca-Cola Company` balance node, the `debit` sheet, update text in `(R1,C1)`, add text to `(R3,C1)` and remove `(R2,C1)`:

```bash
curl -s -XPATCH -H "content-type: application/json" -d "{ \"upsertRecords\": [{\"id\": \"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"updated outgoing company transfer \$2000000\", \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\": \"debit|R3|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"added outgoing company transfer \$3000000\", \"vector\": \"`echo '["debit", "R3", "C1"]' | base64`\"}], \"deleteRecords\": [{\"id\": \"debit|R2|C1\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234%2Fbalance.xlsx/records" | jq
```

### 7. Search the records of the modified node

Search the `Coca-Cola Company` balance node only, check that the records have changed:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234/balance.xlsx", "tags": {}, "strict":true, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```

### 8. Delete nodes

Delete the `Coca-Cola Company` node and all its children nodes and records:

```bash
curl -i -XDELETE "http://localhost:8080/v1/nodes/%2Forgs%2F1234"
```

### 9. List the nodes after deletion

List the organization node children, only the `Ford Motors Company` nodes are expected:

```bash 
curl -s -XGET "http://localhost:8080/v1/nodes?path=/orgs" | jq
```

### 10. Search the records of all the organizations

Search the organization node children, only the `Ford Motors Company` records are expected:

```bash
curl -s -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "tags": {}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search" | jq
```
