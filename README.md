# Dchar - Docker Compose Harness

For easily setting up tests that rely on services in a docker-compose.yml

Wrap you `m.Run()` and `dchar` will take care of spinning up your containers and checking that they're ready before
running your tests.

Have a look in the [./examples](./examples) dir for runnable tests.

## Usage

```go
package my_test

import (
	"os"
	"testing"
	"github.com/byrnedo/dchar"
)

func TestMain(m *testing.M) {

	h := dchar.Harness{
		ProjectName: "dchar-example",
		Services: []dchar.Service{
			{
				Name: "nats",
				// will pull before starting tests
				Pull: true,
				// will wait for nats to listen on localhost:4222
				Waiter: dchar.TcpWaiter{
					Port: 14222,
				},
			},
		},
	}

	c := 0
	// h.Run does
	// - down (removes volumes)
	// - pull (for any configured services with Pull = TRUE)
	// - build
	// - up
	// And when the callback is finished, will run down
	if err := h.Run(context.Background(), func() {
		c = m.Run()
	}); err != nil {
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