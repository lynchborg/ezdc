package logfile_test

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
		Logs:        dctest.FileLogWriter("./logs/docker-compose.log"),
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
