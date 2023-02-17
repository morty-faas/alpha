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
# You have to define the process to run when the agent starts. You can do it by setting the `ALPHA_INVOKE` environment variable. For example: 
export ALPHA_INVOKE="node /tmp/index.js"

go run main.go

# You should see a similar output (your version of node can be different) : 
INFO[0000] invoke : /usr/bin/node /tmp/index.js 
Node v18.12.1 is listening on 0.0.0.0:3000

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

Execute a function with variables through the agent :

WIP



