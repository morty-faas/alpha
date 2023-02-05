# Alpha

**Alpha** is a little agent written in **Go** responsible for running function code on the host where it runs.

Currently, Alpha exposes only a single HTTP endpoint on the port.

## State

The project is a mainly a PoC for now. We will need to refactor some things and add tests to ensure it works as expected. Don't hesitate to open issues or PR to fix / add features. 

## Runtimes

Alpha has a support only the following runtimes for now :

- NodeJS

## Development

**Alpha was written using **Go 1.19**. We recommend you to install at least **Go 1.19** in order to be able to develop the project.** You need to have at least Node.JS installed locally.

Start the agent :

```bash
go run main.go

# You should see a similar output (your version of node can be different) : 
INFO[0000] Runtime initialized                           runtime=node version=v19.4.0
INFO[0000] Executor ready. 1 runtime(s) available for this agent 
INFO[0000] HTTP API Server is listening on 0.0.0.0:3000 
```

Execute a function without variables through the agent : 

```bash
curl -XPOST http://localhost:3000 -d '
{
  "runtime": "node",
  "code": "https://gitlab.com/N4rkos/js-lambda-showcase/-/archive/main/js-lambda-showcase-main.tar.gz"
}'

# You should have something like this (the animal name is random at each execution) : 
{
 "output": "{ message: 'Hello world !', animalName: 'Carnivorous Stinkbug' }\n",
 "process_exit_code": 0
}
```

Execute a function with variables through the agent :

```bash
curl -XPOST http://localhost:3000?name=Alpha -d '
{
  "runtime": "node",
  "code": "https://gitlab.com/N4rkos/js-lambda-showcase/-/archive/main/js-lambda-showcase-main.tar.gz"
}'

# You should have something like this (the animal name is random at each execution) : 
{
 "output": "{ message: 'Hello Alpha !', animalName: 'Carnivorous Stinkbug' }\n",
 "process_exit_code": 0
}
```

The example function code can be found [here](https://gitlab.com/N4rkos/js-lambda-showcase).


