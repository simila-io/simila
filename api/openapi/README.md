# Simila API
This folder contains [simila.yaml](simila.yaml) file in [OpenAPI(Swagger)](https://www.openapis.org/) format for Simila service.

## Conventions
Whoever makes changes in the public API should follow the conventions described here.

### Routes(endpoints)
We are going to follow conventions described [here](https://restfulapi.net/resource-naming/). Which in shorts means the following:
* A route consists of resources which are plural nouns:
    ```
  GET /v1/indexes/...
  GET /v1/formats/:id/...
  ```
* In some cases more than one word may be needed for a route, this case we use kebab style:
```
  POST /v1/users/:id/reset-password/...
```
* A resource name is followed by the resource identifier:
    ```
  GET /v1/indexes/1234
  PUT /v1/indexes/1234
  ```
* Instead of a resource name could be a verb, which means the action. This action can be executed by `POST` method:
```
    POST /v1/search
```
* The resource names and action verbs are lowercases - one word

### Models(objects)
A model describes a structure of the object in JSON format. The fields in the model should be camelCased starting from lowercase. Examples:
```
    "id": ...
    "indexId": ...
    "myCustomField": ...
```

### Collections
Collections are the API routes that can return many records at once, hereafter the agreements are defined on how to implement filtering, searching, ordering, and pagination of the records returned by the collections.

#### Filtering
Filtering allows controlling which records should be included in the API response.

* Filtering by the collection model field values is done by specifying the **{fieldName}={value}** parameters in the API request, for instance:
```
   GET /organizations/og1/users?status=active
```

* Filtering by a list of values of the collection model field is done by specifying multiple **{value}** values for the **{fieldName}** field, for instance:
```
   GET /organizations/og1/users?id=us1&id=us2
```

* Filtering by the range of values of the collection model field is done by specifying the **from{fieldName}={fromValue}** and **to{fieldName}={toValue}** parameters (one or both) for every **{fieldName}** field to be filtered by, for instance:
```
    GET /organizations/og1/users?fromCreatedAt=2022-01-02&toCreatedAt=2022-02-02
```

**NOTE**: Filtering via GET parameters fits best for "non-sensitive" data (e.g. record ID, state, status, timeframe), the other cases should be considered on case-by-case basis to decide how to better implement it.

#### Searching
Searching is a case-insensitive lookup of a prefix or substring (this is defined by the API implementation) within the default fields or the specified fields of the collection model.

* Searching by the values of the default collection model fields is done by specifying the **search={textToSearch}** parameter in the request. In this case, the API implementation implicitly decides values of which fields of the collection model should be checked for the occurrence of **{textToSearch}**, for instance:
```
    GET /organizations/og1/users?search=john
```

* Searching by the values of the explicitly specified fields of the collection model is done by specifying the search fields via the **searchField={fieldName}** parameters (additionally to the **search={textToSearch}** parameter), for instance:
```
    GET /organizations/og1/users?search=john&searchField=firstName&searchField=lastName
```

**NOTE**: Searching via GET parameters may expose certain business and PII information in logs, so it should be used with care and maybe replaced with POST.

#### Ordering
Ordering allows specifying the ordering in which the records are returned in the API response.

* Ordering by the values of the collection model field (by default, the order is ascending) is done by specifying the **orderBy={fieldName}** parameter in the request, for instance:
```
    GET /organizations/og1/users?orderBy=createdAt
```

* Ordering in the descending order is done by specifying the **desc** parameter (additionally to the **orderBy={fieldName}** parameter), for instance:
```
    GET /organizations/og1/users?orderBy=createdAt&desc=true
```

#### Pagination
Pagination allows iterating over the response records page by page rather than fetching them all at once.

* Pagination is done by specifying the **fromPageId** and **limit** parameters in the API request; the API response must contain the **items**, **nextPageId** and **total** fields, for instance:
```
    GET /organizations/og1/users?fromPageId=us3&limit=100
    ...
    {
        "items": [...],
        "nextPageId": "us103",
        "total": 100
    }
```

### Versions
The current version API is `v1`, if we create a new version all routes, models and rules will be described there. We don't guarantee that the conventions are going to be same for other versions.

The changes for the API version is "backward compatible", which means the following - In the case of the API change is needed, any client which works with the version `v1` of the API before the change must work with the version of the API after the change.

**NOTE**: The rule doesn't mean that the client which relies on the change must work with the API before the change has been introduced.

### Incremental changes
As soon as the API is "stabilized", to support the "backward compatibility" any change in the API will be incremental (no deletions or semantical changes), just additions (fields, parameters and routes).

**NOTE**: The API is not "stabilized" yet, so we can make any changes here including the agreements and conventions.

