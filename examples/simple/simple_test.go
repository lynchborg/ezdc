package simple

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/byrnedo/dchar"
)

func TestMain(m *testing.M) {
	fmt.Println("start")

	h := dchar.Harness{
		ProjectName: "dchar-simple",
		Services: []dchar.Service{
			{
				Name: "nats",
				Pull: true,
				Waiter: dchar.TcpWaiter{
					Port: 14222,
				},
			},
		},
	}

	c := 0
	if err := h.Run(context.Background(), func() {
		c = m.Run()
	}); err != nil {
		panic(err)
	}

	os.Exit(c)
}

func TestNothing(t *testing.T) {
	// whoop
	t.Log("NOTHING")
}
