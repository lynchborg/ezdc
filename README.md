# Dchar - Docker Compose Harness

For easily setting up tests that rely on services in a docker-compose.yml

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