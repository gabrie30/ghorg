# Reclone Server Command

The `reclone-server` command starts a server that allows you to trigger ad hoc reclone commands via HTTP requests.

See the [Reclone Command](https://github.com/gabrie30/ghorg#reclone-command) section of the README for details on configuring `reclone.yaml`.

## Usage

```sh
ghorg reclone-server [flags]
```

## Flags

- `--port`: Specify the port on which the server will run. If not specified, the server will use the default port.

## Endpoints

- **`/trigger/reclone`**: Triggers the reclone command. To prevent resource exhaustion, only one request can processed at a time.
  - **Query Parameters**:
    - `cmd`: Optional. Allows you to call a specific reclone, otherwise all reclones are ran.
  - **Responses**:
    - `200 OK`: Command started successfully.
    - `429 Too Many Requests`: Server is currently running a reclone command, you will need to wait until its completed before starting another one.

- **`/stats`**: Returns the statistics of the reclone operations in JSON format. `GHORG_STATS_ENABLED=true` or `--stats-enabled` must be set to work.
  - **Responses**:
    - `200 OK`: Statistics returned successfully.
    - `428 Precondition required`: Ghorg stats is not enabled.
    - `500 Internal Server Error`: Unable to read the statistics file.

- **`/health`**: Health check endpoint.
  - **Responses**:
    - `200 OK`: Server is healthy.

## Examples

Starting the server. The default port is `8080` but you can optionally start the server on different port using the `--port` flag:

```sh
ghorg reclone-server
```

Trigger reclone command, this will run all cmds defined in your `reclone.yaml`:

```sh
curl "http://localhost:8080/trigger/reclone"
```

Trigger a specific reclone command:

```sh
curl "http://localhost:8080/trigger/reclone?cmd=your-reclone-command"
```

Get the statistics:

```sh
curl "http://localhost:8080/stats"
```

Check the server health:

```sh
curl "http://localhost:8080/health"
```
