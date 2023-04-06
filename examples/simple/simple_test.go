package simple

import (
	"context"
	"github.com/lynchborg/ezdc"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	h := ezdc.Harness{
		ProjectName: "ezdc-simple",
		Services: []ezdc.Service{
			{
				Name: "nats",
				Pull: true,
				Waiter: ezdc.TcpWaiter{
					Port: 14222,
				},
			},
			{
				Name: "echo",
				Waiter: ezdc.HttpWaiter{
					Port: 5678,
				},
			},
		},
	}

	c, err := h.Run(context.Background(), m.Run)
	if err != nil {
		panic(err)
	}

	os.Exit(c)
}

func TestNothing(t *testing.T) {
	// whoop
	t.Log("NOTHING")
}
