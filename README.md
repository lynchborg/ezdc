# EZDC - Easy Testing With Docker Compose
[![Go Reference](https://pkg.go.dev/badge/github.com/lynchborg/ezdc.svg)](https://pkg.go.dev/github.com/lynchborg/ezdc)
[![Go Coverage](https://github.com/lynchborg/ezdc/wiki/coverage.svg)](https://raw.githack.com/wiki/lynchborg/ezdc/coverage.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/lynchborg/ezdc)](https://goreportcard.com/report/github.com/lynchborg/ezdc)

For easily setting up tests that rely on services in a docker-compose.yml

Do you want your tests to be setup with:

- `docker compose pull`
- `docker compose build`
- `docker compose up`
- Some logic to know containers are ready
- YOUR TESTS HERE

  Followed by
- `docker compose down` ?

Yes? Then we've got you covered. No? Make a P.R. ⌨️ ❤️

Wrap you `m.Run()` and `ezdc` will take care of spinning up your containers and checking that they're ready before
running your tests.

Have a look in the [./examples](./examples) dir for runnable tests.

## Usage

```go
package my_test

import (
	"os"
	"testing"
	"github.com/byrnedo/ezdc"
)

func TestMain(m *testing.M) {

	h := ezdc.Harness{
		ProjectName: "ezdc-example",
		Services: []ezdc.Service{
			{
				Name: "nats",
				// will pull before starting tests
				Pull: true,
				// will wait for nats to listen on localhost:4222
				Waiter: ezdc.TcpWaiter{
					Port: 4222,
				},
			},
		},
	}

	// h.Run does
	// - down (removes volumes)
	// - pull (for any configured services with Pull = TRUE)
	// - build
	// - up
	// And when the callback is finished, will run down
	c, err := h.Run(context.Background(), m.Run)
	if err != nil {
		panic(err)
	}

	os.Exit(c)
}
```

## Configuration

### Harness

| Setting     |                                                                 |
|-------------|-----------------------------------------------------------------|
| ProjectName | The project name to give to docker-compose. Required            | 
| File        | Path to docker-compose file. Defaults to ./docker-compose.yml   |
| Services    | Configuration of services. Optional.                            |
| Logs        | Where to send the docker-compose output. Defaults to os.Stdout. |

### Services

| Setting |                                                                                                                   |
|---------|-------------------------------------------------------------------------------------------------------------------|
| Name    | Name of the service. Doesn't necessarily have to match any in the docker compose file, but does if `Pull` is TRUE |
| Pull    | Pull the image before running tests. Default is false.                                                            |
| Waiter  | Configures how to declare your service 'ready'. Optional.                                                         |

## Waiters

Currently supports

- Http (`HttpWaiter`)
- Tcp (`TcpWaiter`)
