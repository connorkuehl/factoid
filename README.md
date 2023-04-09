# factoid

A RESTful API for sharing fun facts.

## API reference

### Fact

#### Get a random fact

To get a random fact, send a GET request to `/v1/fact/rand`.

Example:

```console
curl http://factoid.example.com/v1/fact/rand
```

Response [HTTP 200]: A JSON object whose "fact" field contains a
random fact.

```json
{
  "fact": {
    "id": 36,
    "created_at": "2023-02-26T16:51:22Z",
    "updated_at": "2023-02-26T16:51:22Z",
    "content": "It looks like you know how to get a random fact!",
    "source": "factoid's README"
  }
}
```

#### Get a fact

To get a fact whose ID is already known to you, send a GET request to `/v1/fact/:id`.

Example:

```console
curl -s http://factoid.example.com/v1/fact/36
```

Response [HTTP 200]: A JSON object whose "fact" field contains the specified
fact.

```json
{
  "fact": {
    "id": 36,
    "created_at": "2023-02-26T16:51:22Z",
    "updated_at": "2023-02-26T16:51:22Z",
    "content": "It looks like you know how to get a random fact!",
    "source": "factoid's README"
  }
}
```

Response [HTTP 400]: A JSON object whose "error" field describes what is wrong
with the request.

```json
{
  "error": "id must be an integer or 'rand'"
}
```

Response [HTTP 404]: A JSON object whose "error" field indicates a fact with
the specified ID was not found.

```json
{
  "error": "not found"
}
```

#### Get all facts

To get all facts, send a GET request to `/v1/facts`.

Example:

```console
curl -s http://factoid.example.com/v1/facts
```

Response [HTTP 200]: A JSON object whose "facts" field contains an
array of all the facts known to the server.

```json
{
  "facts": [
    {
      "id": 2,
      "created_at": "2023-02-26T16:51:21Z",
      "updated_at": "2023-02-26T16:51:21Z",
      "content": "Some fact",
      "source": "A twitter account"
    },
    {
      "id": 36,
      "created_at": "2023-02-26T16:51:22Z",
      "updated_at": "2023-02-26T16:51:22Z",
      "content": "It looks like you know how to get a random fact!",
      "source": "factoid's README"
    }
  ]
}
```

#### Create a fact

To create a fact, send a POST request to `/v1/facts`. The server expects
a JSON payload in the body of your request.

Example:

```console
curl -s -d '{"content": "A new fact", "source": "A README document"}' http://factoid.example.com/v1/facts
```

Response [HTTP 201]: The newly created fact object.

```json
{
  "fact": {
    "id": 38,
    "created_at": "2023-02-26T17:21:36Z",
    "updated_at": "2023-02-26T17:21:36Z",
    "content": "A new fact",
    "source": "A README document"
  }
}
```

Response [HTTP 400]: A JSON object whose error field describes what is
wrong with the request.

```json
{
  "error": "content field missing or blank"
}
```

#### Delete a fact

To delete a fact, send a DELETE request to `/v1/fact/:id`.

Note: this is not yet implemented.

