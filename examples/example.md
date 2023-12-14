### 1. Create formats for records

```bash
# Create format for organization meta (e.g. name)
curl -i -XPOST -H "content-type: application/json" -d '{"name": "organizationsMeta", "basis": ["table", "id", "column"]}' "http://localhost:8080/v1/formats"

# Create format for spreadsheet meta (e.g. filename)
curl -i -XPOST -H "content-type: application/json" -d '{"name": "spreadsheetsMeta", "basis": ["path", "filename"]}' "http://localhost:8080/v1/formats"

# Create format for spreadsheet data (e.g. data from cells)
curl -i -XPOST -H "content-type: application/json" -d '{"name": "spreadsheetsData", "basis": ["sheet", "row", "col"]}' "http://localhost:8080/v1/formats"
```

### 2. List the created formats

```bash
# List created formats
curl -i -XGET "http://localhost:8080/v1/formats"
```

### 3. Create searchable records (and corresponding nodes)

```bash
# Make "Coca-Cola Company" organization name searchable for everyone (via tags)
curl -i -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"folder\", \"tags\": {\"public\": \"true\"}, \"records\": [{\"id\":\"organizations|1234|name\", \"format\": \"organizationsMeta\", \"segment\": \"Coca-Cola Company\", \"rankMultiplier\": 2.0, \"vector\": \"`echo '["organizations", "1234", "name"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234/records"

# Make "Coca-Cola Company" organization balance spreadsheet searchable (both meta and data) for internal use only (via tags)
curl -i -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"document\", \"tags\": {\"public\": \"false\"}, \"records\": [{\"id\":\"/spreadsheets/2023|balance.xlsx|name\", \"format\": \"spreadsheetsMeta\", \"segment\": \"company balance 2023\", \"rankMultiplier\": 1.5, \"vector\": \"`echo '["/spreadsheets/2023", "balance.xlsx"]' | base64`\"}, {\"id\":\"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$100\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\":\"debit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$200\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R2", "C1"]' | base64`\"}, {\"id\":\"credit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$1000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R1", "C1"]' | base64`\"}, {\"id\":\"credit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$2000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R2", "C1"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234%2Fbalance.xlsx/records"

# Make "Ford Motors Company" organization name searchable for everyone (via tags)
curl -i -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"folder\", \"tags\": {\"public\": \"true\"}, \"records\": [{\"id\":\"organizations|5678|name\", \"format\": \"organizationsMeta\", \"segment\": \"Ford Motors Company\", \"rankMultiplier\": 2.0, \"vector\": \"`echo '["organizations", "5678", "name"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F5678/records"

# Make "Ford Motors Company" organization balance spreadsheet searchable (both meta and data) for everyone (via tags)
curl -i -XPOST -H "content-type: application/json" -d "{\"nodeType\": \"document\", \"tags\": {\"public\": \"true\"}, \"records\": [{\"id\":\"/spreadsheets/2023|balance.xlsx|name\", \"format\": \"spreadsheetsMeta\", \"segment\": \"company balance 2023\", \"rankMultiplier\": 1.5, \"vector\": \"`echo '["/spreadsheets/2023", "balance.xlsx"]' | base64`\"}, {\"id\":\"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$300\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\":\"debit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"outgoing company transfer \$600\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["debit", "R2", "C1"]' | base64`\"}, {\"id\":\"credit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$3000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R1", "C1"]' | base64`\"}, {\"id\":\"credit|R2|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"incoming company transfer \$6000\", \"rankMultiplier\": 1.0, \"vector\": \"`echo '["credit", "R2", "C1"]' | base64`\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F5678%2Fbalance.xlsx/records"
```

### 4. List the created nodes

```bash
# List organization node children
curl -i -XGET "http://localhost:8080/v1/nodes?path=/orgs"

# List "Coca-Cola Company" node children
curl -i -XGET "http://localhost:8080/v1/nodes?path=/orgs/1234"

# List "Ford Motors Company" node children
curl -i -XGET "http://localhost:8080/v1/nodes?path=/orgs/5678"
```

### 5. Search created records by specifying different paths, tags etc.

```bash
# Search organization node children, return most relevant record per node only
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "tags": {}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search"

# Search organization node children with filter via tags, return most relevant record per node only
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs", "tags": {"shared":"true"}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search"

# Search "Coca-Cola Company" all node children, return most relevant record per node only
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234", "tags": {}, "strict":false, "offset":0, "limit":100}' "http://localhost:8080/v1/search"

# Search "Coca-Cola Company" balance node only, return all matched records for the node
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234/balance.xlsx", "tags": {}, "strict":true, "offset":0, "limit":100}' "http://localhost:8080/v1/search"
```

### 6. Add, update and delete searchable records of a node

```bash
# Modify "Coca-Cola Company" balance node, "debit" sheet, update text in {R1,C1}, add text to {R3,C1} and remove {R2,C1}
curl -i -XPATCH -H "content-type: application/json" -d "{ \"upsertRecords\": [{\"id\": \"debit|R1|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"updated outgoing company transfer \$2000000\", \"vector\": \"`echo '["debit", "R1", "C1"]' | base64`\"}, {\"id\": \"debit|R3|C1\", \"format\": \"spreadsheetsData\", \"segment\": \"added outgoing company transfer \$3000000\", \"vector\": \"`echo '["debit", "R3", "C1"]' | base64`\"}], \"deleteRecords\": [{\"id\": \"debit|R2|C1\"}]}" "http://localhost:8080/v1/nodes/%2Forgs%2F1234%2Fbalance.xlsx/records"
```

### 7. Search the records of the modified node

```bash
# Search "Coca-Cola Company" balance node only, the records must have changed
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234/balance.xlsx", "tags": {}, "strict":true, "offset":0, "limit":100}' "http://localhost:8080/v1/search"
```

### 8. Delete nodes

```bash
# Delete "Coca-Cola Company" balance node
curl -i -XDELETE "http://localhost:8080/v1/nodes/%2Forgs%2F1234%2Fbalance.xlsx"

# Delete "Coca-Cola Company" node and all its nodes and records
curl -i -XDELETE "http://localhost:8080/v1/nodes/%2Forgs%2F1234"
```

### 9. List the nodes after deletion

```bash
# List organization node children, only "Ford Motors Company" nodes are expected
curl -i -XGET "http://localhost:8080/v1/nodes?path=/orgs"

# List "Coca-Cola Company" node children, "not found" is expected
curl -i -XGET "http://localhost:8080/v1/nodes?path=/orgs/1234"
```

### 10. Search the records of the deleted node

```bash
# Search "Coca-Cola Company" balance node, 0 records must be returned
curl -i -XPOST -H "content-type: application/json" -d '{"text": "company", "path":"/orgs/1234/balance.xlsx", "tags": {}, "strict":true, "offset":0, "limit":100}' "http://localhost:8080/v1/search"
```