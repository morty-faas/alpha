# Alpha

**Alpha** is a little agent written in **Go** responsible for running function code on the host where it runs.

Currently, Alpha exposes only two HTTP endpoints on the port 8080 :

- `POST http://localhost:8080` : Call the function, see the [development](#development) section.
- `GET http://localhost:8080/_/health` : Get the health status of the agent.

## State

The project is a mainly a PoC for now. We will need to refactor some things and add tests to ensure it works as expected. Don't hesitate to open issues or PR to fix / add features.

## Development

**Alpha was written using **Go 1.19**. We recommend you to install at least **Go 1.19** in order to be able to develop the project.** You need to have at least Node.JS installed locally.

Set up the environment file :

```bash
cp .env.example .env

# Or either set the `ALPHA_INVOKE` environment variable,
# to define the process to run when the agent starts :
export ALPHA_INVOKE="node /tmp/index.js"
```

Start the agent :

```bash
go run *.go

# You should see a similar output (your process can be different)
INFO[0000] Started process node /runtimes/template/node-19/index.js (pid=41322)
INFO[0000] Alpha server listening on 0.0.0.0:8080
downstream: Node v19.4.0 is listening on 0.0.0.0:3000
```

Alpha is now listening on the port 8080 and will forward http calls to port 3000.

Execute a function without variables through the agent :

```bash
curl -XPOST http://localhost:8080

# You should receive the response from the function :
My first function !

# You can see the logs from the function in the agent's logs :
Random animal name : Curious Seahorse

```

## Configuration

The application support both `.env` file and environment variables. All the environment variables located in the `.env` file can be overriden with "classic" environment variables using : `export VARIABLE=value`.

The supported environment variables are the following :

- `ALPHA_INVOKE` (**required**): The process to run when the agent starts
- `ALPHA_REMOTE` : The address where the child process listens on (default is `http://127.0.0.1:3000`)
- `ALPHA_PORT` : The port on which the application will run (default : `8080`)
- `ALPHA_LOG_LEVEL` : The logging level (default is `4`, which is `Info`)

Execute a function with variables through the agent :

WIP
