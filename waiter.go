package ezdc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// Waiter should implement Wait to return nil once the service is ready
type Waiter interface {
	Wait(context.Context) error
}

// TcpWaiter checks if a tcp connection can be established
type TcpWaiter struct {
	Interval time.Duration
	Timeout  time.Duration
	Port     int
}

func (tw TcpWaiter) host() string {
	return fmt.Sprintf("localhost:%d", tw.Port)
}

func (tw TcpWaiter) Wait(ctx context.Context) error {

	interval := tw.Interval
	if interval == 0 {
		interval = 500 * time.Millisecond
	}
	timeout := tw.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	for {

		d := net.Dialer{
			Timeout: timeout,
		}
		var c net.Conn
		c, err := d.DialContext(ctx, "tcp", tw.host())
		if err == nil {
			_ = c.Close()
			return nil
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		info(fmt.Sprintf(" failed to connect: '%s'", err))
		time.Sleep(interval)
	}

}

// HttpWaiter ensures a healthy status code is received from a http endpoint
type HttpWaiter struct {
	Interval       time.Duration
	RequestTimeout time.Duration
	Port           int
	Path           string
	ReadyStatus    []int
}

func (hw HttpWaiter) url() string {
	path := hw.Path
	if !strings.HasPrefix(path, "/") {
		path += "/" + path
	}
	return fmt.Sprintf("http://localhost:%d%s", hw.Port, path)
}

func (hw HttpWaiter) Wait(ctx context.Context) error {
	readyStatus := hw.ReadyStatus
	if len(readyStatus) == 0 {
		readyStatus = []int{200, 201, 202, 204}
	}

	interval := hw.Interval
	if interval == 0 {
		interval = 500 * time.Millisecond
	}
	requestTimeout := hw.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 2 * time.Second
	}
	u := hw.url()

	for {
		var (
			res *http.Response
			err error
		)
		func() {
			ctx, cncl := context.WithTimeout(ctx, requestTimeout)
			defer cncl()
			req, _ := http.NewRequest("GET", u, nil)
			req = req.WithContext(ctx)
			res, err = http.DefaultClient.Do(req)
		}()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			info(fmt.Sprintf("failed to connect: err='%s'", err))
			//_, _ = fmt.Fprintf(os.Stderr, "%s - %s\n", name, err)
		} else if find(readyStatus, res.StatusCode) {
			return nil
		} else {
			info(fmt.Sprintf("failed to connect: status='%d'", res.StatusCode))
		}
		time.Sleep(interval)
	}

}
