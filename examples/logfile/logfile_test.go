package logfile_test

import (
	"context"
	"os"
	"testing"

	"github.com/lynchborg/ezdc"
)

func TestMain(m *testing.M) {

	h := ezdc.Harness{
		ProjectName: "ezdc-logfile",
		Logs:        ezdc.FileLogWriter("./logs/docker-compose.log"),
		Services: []ezdc.Service{
			{
				Name: "nats",
				Pull: true,
				Waiter: ezdc.TcpWaiter{
					Port: 24222,
				},
			},
		},
	}

	c, err := h.Run(context.Background(), m.Run)
	if err != nil {
		// something went wrong with the test harness
		panic(err)
	}

	os.Exit(c)
}

func TestNothing(t *testing.T) {
	// whoop
	t.Log("NOTHING")
}
