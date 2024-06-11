# API documentation

## Create a server

```http
POST /api/config/server
Content-Type: application/json

{
  "name": "server-name",
  "cmd": "command to execute",
  "cwd": "directory to execute command in",
  "use_direnv": true,
  "env": {
    "ENV_VAR": "value"
  }
}
```

## Delete a server

```http
DELETE /api/config/server/:name
```

## Get all servers configuration

```http
GET /api/config/server
```

## Start a server

```http
POST /api/runner/:name
```

## Stop a server

```http
DELETE /api/runner/:name
```

## Fetch overview of state of all servers

```http
GET /api/state
```

## Fetch state and logs of of a specific server

```http
GET /api/state/:name
```

## Websocket to get real-time state updates

```http
GET /api/ws
```

This will return the state of all servers and will update in real-time.

If the client sends a message in the format of `{"name": "server-name"}`,
it will also return the logs of that server along side the overview state
of all the servers.
