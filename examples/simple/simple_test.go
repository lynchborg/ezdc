package simple

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/byrnedo/dctest"
)

func TestMain(m *testing.M) {
	fmt.Println("start")

	h := dctest.Harness{
		ProjectName: "dctest-simple",
		Services: []dctest.Service{
			{
				Name: "nats",
				Pull: true,
				Waiter: dctest.TcpWaiter{
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
