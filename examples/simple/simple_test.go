package simple

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/byrnedo/ezdc"
)

func TestMain(m *testing.M) {
	fmt.Println("start")

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
	time.Sleep(10 * time.Second)
	t.Log("NOTHING")
}
