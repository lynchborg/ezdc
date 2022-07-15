package logfile_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/byrnedo/ezdc"
)

func TestMain(m *testing.M) {
	fmt.Println("start")

	h := ezdc.Harness{
		ProjectName: "ezdc-simple",
		Logs:        ezdc.FileLogWriter("./logs/docker-compose.log"),
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

	c := 0
	if err := h.Run(context.Background(), func() {
		// NOTE: don't call os.Exit in here
		// IF you do, the test main will hang indefinitely
		c = m.Run()
	}); err != nil {
		// something went wrong with the test harness
		panic(err)
	}

	os.Exit(c)
}

func TestNothing(t *testing.T) {
	// whoop
	t.Log("NOTHING")
}
