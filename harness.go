package ezdc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"time"
)

func infoLog(msg string) {
	log.Println("#### EZDC #### INFO " + msg)
}

func errorLog(msg string) {
	log.Println("#### EZDC #### ERROR " + msg)
}

// Service configures options for a service defined in the docker compose file
type Service struct {
	Name   string
	Pull   bool   // pull before starting tests
	Waiter Waiter // optional, how to wait for service to be ready
}

// FileLogWriter utility to open a file for logging the docker compose output
func FileLogWriter(fileName string) *os.File {
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open log file %s: %w", fileName, err))
	}
	_ = logFile.Truncate(0)
	_, _ = logFile.Seek(0, 0)
	return logFile
}

type Harness struct {
	ProjectName   string    // name for the compose project
	File          string    // path to the docker compose file
	Services      []Service // configuration for services
	Logs          io.Writer // where to send the docker compose logs
	termSig       chan os.Signal
	cc            cc
	cleanerUppers []func(context.Context)
}

// Run is the entrypoint for running your testing.M.
//
// func TestMain(m *testing.M) {
//     h := Harness{.....} // configure
//
//     exitCode, err := h.Run(context.Background(), m.Run)
//     if err != nil {
//         panic(err)
//     }
//     os.Exit(exitCode)
// }
func (h *Harness) Run(ctx context.Context, f func() int) (int, error) {

	h.termSig = make(chan os.Signal)

	h.cc.project = h.ProjectName
	h.cc.file = h.File

	go func() {
		<-h.termSig
		h.cleanup(10 * time.Second)
		os.Exit(1)
	}()
	signal.Notify(h.termSig, os.Interrupt)

	if err := h.startDcServices(ctx); err != nil {
		return 1, err
	}

	if err := h.waitForServices(ctx); err != nil {
		return 1, err
	}

	infoLog("services ready")

	defer func() {
		h.cleanup(10 * time.Second)
	}()
	return f(), nil
}

func (h Harness) startDcServices(ctx context.Context) error {
	_ = h.cc.down(ctx)

	toPull := gmap(
		filter(h.Services, func(s Service) bool {
			return s.Pull
		},
		), func(s Service) string {
			return s.Name
		})

	if err := h.cc.pull(ctx, toPull...); err != nil {
		return fmt.Errorf("error pulling: %w", err)
	}

	if err := h.cc.build(ctx); err != nil {
		return fmt.Errorf("error building: %w", err)
	}

	var buf bytes.Buffer
	upCmd := h.cc.cmd(ctx, "up")
	if h.Logs == nil {
		h.Logs = os.Stdout
	}

	multiWriter := io.MultiWriter(h.Logs, &buf)
	upCmd.Stdout = h.Logs
	upCmd.Stderr = multiWriter

	if err := upCmd.Start(); err != nil {
		return err
	}
	return nil
}

func (h Harness) waitForServices(ctx context.Context) error {

	toWaitFor := filter(h.Services, func(s Service) bool {
		return s.Waiter != nil
	})

	for _, svc := range toWaitFor {
		infoLog(fmt.Sprintf("waiting for '%s'...\n", svc.Name))
		if err := svc.Waiter.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// CleanupFunc registers a function to be run before stopping the docker compose services
func (h *Harness) CleanupFunc(f func(context.Context)) {
	h.cleanerUppers = append(h.cleanerUppers, f)
}

func (h Harness) cleanup(timeout time.Duration) {
	infoLog("cleaning up")

	ctx, cncl := context.WithTimeout(context.Background(), timeout)
	defer cncl()

	for _, f := range h.cleanerUppers {
		f(ctx)
	}

	if err := h.cc.down(ctx); err != nil {
		errorLog(fmt.Sprintf("failed to run 'down': %s", err))
	}
}
