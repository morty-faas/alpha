# Alpha

**Alpha** is a little agent written in **Go** responsible for running function code on the host where it runs.

Currently, Alpha exposes only two HTTP endpoints on the port 8080 :

- `POST http://localhost:8080` : Call the function, see the [development](#development) section.
- `GET http://localhost:8080/_/health` : Get the health status of the agent.

## State

The project is a mainly a PoC for now. We will need to refactor some things and add tests to ensure it works as expected. Don't hesitate to open issues or PR to fix / add features.

## Development

**Alpha was written using **Go 1.19**. We recommend you to install at least **Go 1.19** in order to be able to develop the project.** You need to have at least Node.JS installed locally.

Start the agent :

```bash
export ALPHA_PROCESS_COMMAND="node /runtimes/template/node-19/index.js"
go run *.go

# You should see a similar output (your process can be different)
INFO[0000] Started process node /runtimes/template/node-19/index.js (pid=41322)
INFO[0000] Alpha server listening on 0.0.0.0:8080
```

Alpha is now listening on the port 8080 and will forward http calls to port 3000.

Execute a function without variables through the agent :

```bash
curl http://localhost:8080
```

You should see a similar output (values can be different depending on what you're launching. `payload` will be the return of your function) : 
```json
{
  "payload": "Hello, world !",
  "process_metadata": {
    "execution_time_ms": 149,
    "logs": [
      "Sending request to https://jsonplaceholder.typicode.com/posts/1"
    ]
  }
}
```

## Configuration

The application support both `yaml` config file and environment variables.

Below the reference of the configuration file: 

```yaml
server:
  port: 8080
  logLevel: INFO

process:
  command: "node /tmp/index.js"
  downstreamUrl: "http://127.0.0.1:3000"
```

Every configuration key can be overridden through environment variables. For example to override the `server.port` value, export the following environment variable `ALPHA_SERVER_PORT=8080`.
