package ezdc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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

type crasher interface {
	Crash(exitCode int)
}

type exitCrasher struct{}

func (e exitCrasher) Crash(exitCode int) {
	os.Exit(exitCode)
}

type Harness struct {
	ProjectName   string    // name for the compose project
	File          string    // path to the docker compose file
	Services      []Service // configuration for services
	Logs          io.Writer // where to send the docker compose logs
	termSig       chan os.Signal
	cc            ComposeFace
	cleanerUppers []func(context.Context)
	crasher       crasher
}

// Run is the entrypoint for running your testing.M.
//
//	func TestMain(m *testing.M) {
//	    h := Harness{.....} // configure
//
//	    exitCode, err := h.Run(context.Background(), m.Run)
//	    if err != nil {
//	        panic(err)
//	    }
//	    os.Exit(exitCode)
//	}
func (h *Harness) Run(ctx context.Context, f func() int) (code int, err error) {

	ctx, cncl := context.WithCancel(ctx)
	defer cncl()
	h.termSig = make(chan os.Signal)
	if h.cc == nil {
		h.cc = DefaultComposeCmd{
			project: h.ProjectName,
			file:    h.File,
		}
	}
	if h.crasher == nil {
		h.crasher = exitCrasher{}
	}

	go func() {
		<-h.termSig
		infoLog("got os Signal")
		cncl()
		h.cleanup(10 * time.Second)
		h.crasher.Crash(1)
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
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("panic: %s", panicErr)
		}
		h.cleanup(10 * time.Second)
	}()
	return f(), nil
}
func (h Harness) withLogs(cmd *exec.Cmd) (*exec.Cmd, *bytes.Buffer) {

	if h.Logs == nil {
		h.Logs = os.Stdout
	}

	buf := &bytes.Buffer{}
	cmd.Stdout = h.Logs
	cmd.Stderr = io.MultiWriter(h.Logs, buf)
	return cmd, buf
}

func (h Harness) startDcServices(ctx context.Context) error {

	infoLog("cleaning up lingering resources")
	cmd, _ := h.withLogs(h.cc.Down(ctx))
	_ = cmd.Run()

	toPull := gmap(
		filter(h.Services, func(s Service) bool {
			return s.Pull
		},
		), func(s Service) string {
			return s.Name
		})

	if len(toPull) > 0 {
		cmd, errBuf := h.withLogs(h.cc.Pull(ctx, toPull...))
		infoLog("pulling")
		if err := cmd.Run(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, errBuf.String())
			return fmt.Errorf("error pulling: %w", err)
		}
	}

	cmd, errBuf := h.withLogs(h.cc.Build(ctx))

	infoLog("building")
	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errBuf.String())
		return fmt.Errorf("error building: %w", err)
	}

	cmd, errBuf = h.withLogs(h.cc.Up(ctx))

	infoLog("starting")
	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errBuf.String())
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

	cmd, errBuf := h.withLogs(h.cc.Down(ctx))

	if err := cmd.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errBuf.String())
		errorLog(fmt.Sprintf("failed to run 'down': %s", err))
	}
}
