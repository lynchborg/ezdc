package ezdc

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

type mockCompose struct {
	up    func(ctx context.Context) *exec.Cmd
	down  func(ctx context.Context) *exec.Cmd
	build func(ctx context.Context) *exec.Cmd
	stop  func(ctx context.Context) *exec.Cmd
	pull  func(ctx context.Context, svc ...string) *exec.Cmd
}

func (m mockCompose) defaultImpl(ctx context.Context, name string) *exec.Cmd {
	return exec.CommandContext(ctx, "echo", name)
}

func (m mockCompose) Up(ctx context.Context) *exec.Cmd {
	if m.up == nil {
		return m.defaultImpl(ctx, "up")
	}
	return m.up(ctx)
}

func (m mockCompose) Down(ctx context.Context) *exec.Cmd {
	if m.down == nil {
		return m.defaultImpl(ctx, "down")
	}
	return m.down(ctx)
}

func (m mockCompose) Stop(ctx context.Context) *exec.Cmd {
	if m.stop == nil {
		return m.defaultImpl(ctx, "stop")
	}
	return m.stop(ctx)
}

func (m mockCompose) Pull(ctx context.Context, svc ...string) *exec.Cmd {
	if m.pull == nil {
		return m.defaultImpl(ctx, "pull")
	}
	return m.pull(ctx, svc...)
}

func (m mockCompose) Build(ctx context.Context) *exec.Cmd {
	if m.build == nil {
		return m.defaultImpl(ctx, "build")
	}
	return m.build(ctx)
}

type mockWaiter struct {
	duration time.Duration
}

func (m mockWaiter) Wait(ctx context.Context) error {
	timer := time.NewTimer(m.duration)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil
		case <-ctx.Done():
			return nil
		}
	}
}

type mockCrasher struct {
	crashed bool
	code    int
}

func (m *mockCrasher) Crash(code int) {
	m.crashed = true
	m.code = code
}
func TestShouldCleanupUponKillSignal(t *testing.T) {
	mockCrash := &mockCrasher{}

	h := Harness{
		ProjectName: t.Name(),
		File:        "./whatever.yml",
		Services: []Service{
			{
				Name:   "foo",
				Waiter: mockWaiter{10 * time.Second},
			},
		},
		crasher: mockCrash,
	}

	h.cc = mockCompose{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = h.Run(context.Background(), func() int {
			return 0
		})
		wg.Done()
	}()
	time.Sleep(100 * time.Millisecond)

	t.Log("sending interrupt")
	h.termSig <- os.Interrupt

	wg.Wait()

	if mockCrash.crashed == false {
		t.Fatal("should have crashed")
	}

	if mockCrash.code != 1 {
		t.Fatalf("should have had exit code %d, had %d", 1, mockCrash.code)
	}

}

func TestShouldErrIfPanicDuringRun(t *testing.T) {

	h := Harness{
		ProjectName: t.Name(),
		File:        "./whatever.yml",
		cc:          mockCompose{},
	}
	_, err := h.Run(context.Background(), func() int {
		panic("foo")
	})

	if err == nil {
		t.Fatal("should return error")
	}
}

func TestShouldRunOurCleanupFunc(t *testing.T) {

	h := Harness{
		ProjectName: t.Name(),
		File:        "./whatever.yml",
		cc:          mockCompose{},
	}
	cleaned := false
	h.CleanupFunc(func(ctx context.Context) {
		t.Log("custom cleanup")
		cleaned = true
	})

	_, err := h.Run(context.Background(), func() int {
		return 0
	})

	if err != nil {
		t.Fatal(err)
	}

	if cleaned != true {
		t.Fatal("cleanup func wasnt run")
	}
}

func TestShouldErrorIfDockerComposeCommandCantStart(t *testing.T) {

	h := Harness{
		ProjectName: t.Name(),
		File:        "./whatever.yml",
	}

	mc := mockCompose{}
	mc.up = func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "assumably-nonexisting-command")
	}
	h.cc = mc

	_, err := h.Run(context.Background(), func() int {
		return 0
	})

	if err == nil {
		t.Fatalf("should have errored")
	}
	t.Log(err)
}
