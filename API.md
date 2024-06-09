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
