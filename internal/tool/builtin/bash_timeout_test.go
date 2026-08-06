package builtin

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"inx/internal/sandbox"
)

func TestBashIdleTimeoutConstants(t *testing.T) {
	// Verify that the idle timeout constants are set correctly.
	if idleTimeoutInitial != 10*time.Minute {
		t.Fatalf("expected idleTimeoutInitial 10m, got %v", idleTimeoutInitial)
	}
	if idleTimeoutExtend != 2*time.Minute {
		t.Fatalf("expected idleTimeoutExtend 2m, got %v", idleTimeoutExtend)
	}
}

func TestIdleTimeoutManagerInitialTimeout(t *testing.T) {
	const initial = 200 * time.Millisecond
	ctx, m := newIdleTimeoutManager(context.Background(), initial, 100*time.Millisecond)
	defer m.cancel(context.Canceled)

	select {
	case <-ctx.Done():
		if !errors.Is(context.Cause(ctx), errBashIdleTimeout) {
			t.Fatalf("cause = %v, want errBashIdleTimeout", context.Cause(ctx))
		}
	case <-time.After(initial + 300*time.Millisecond):
		t.Fatal("context not cancelled by the initial deadline")
	}
}

func TestIdleTimeoutManagerActivityExtendsDeadline(t *testing.T) {
	// Combined-command scenario "short; echo sep; long": the separator output
	// at ~150ms must push the absolute deadline out (initial + extend), not
	// collapse it to the extension window measured from the activity (which
	// would kill a silent long tail ~450ms in).
	const initial = 600 * time.Millisecond
	const extend = 300 * time.Millisecond
	ctx, m := newIdleTimeoutManager(context.Background(), initial, extend)
	defer m.cancel(context.Canceled)

	time.Sleep(150 * time.Millisecond)
	m.reset()

	select {
	case <-ctx.Done():
		t.Fatalf("context cancelled too early after activity: %v", context.Cause(ctx))
	case <-time.After(initial + 100*time.Millisecond):
		// Still alive past the original deadline: the extension preserved it.
	}

	select {
	case <-ctx.Done():
		if !errors.Is(context.Cause(ctx), errBashIdleTimeout) {
			t.Fatalf("cause = %v, want errBashIdleTimeout", context.Cause(ctx))
		}
	case <-time.After(extend + 400*time.Millisecond):
		t.Fatal("context not cancelled by the extended deadline")
	}
}

func TestIdleTimeoutManagerExtensionsAccumulate(t *testing.T) {
	const initial = 600 * time.Millisecond
	const extend = 300 * time.Millisecond
	ctx, m := newIdleTimeoutManager(context.Background(), initial, extend)
	defer m.cancel(context.Canceled)

	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		m.reset()
	}

	select {
	case <-ctx.Done():
		t.Fatalf("context cancelled too early after repeated activity: %v", context.Cause(ctx))
	case <-time.After(initial + extend + 200*time.Millisecond):
		// 1100ms elapsed; three extensions push the deadline to 1500ms.
	}

	select {
	case <-ctx.Done():
		if !errors.Is(context.Cause(ctx), errBashIdleTimeout) {
			t.Fatalf("cause = %v, want errBashIdleTimeout", context.Cause(ctx))
		}
	case <-time.After(initial + 3*extend + 200*time.Millisecond):
		t.Fatal("context not cancelled after the accumulated deadline passed")
	}
}

func TestBashIdleTimeoutWithActivity(t *testing.T) {
	sh := sandbox.ResolveShell("", "", nil)
	// A command that outputs periodically should not timeout within the test duration.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	// Command that outputs periodically
	cmdStr := "echo result-1; sleep 0.5; echo result-2; sleep 0.5; echo result-3"
	out, err := (bash{shell: sh}).Execute(ctx, argsJSON(t, map[string]any{"command": cmdStr}))
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("command with activity failed: %v (out=%q)", err, out)
	}
	if !strings.Contains(out, "result-3") {
		t.Fatalf("output = %q, want 'result-3'", out)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("command returned too slowly: %v", elapsed)
	}
}

func TestNormalizeBashRunErrorAllowsPreservedWaitDelay(t *testing.T) {
	if err := normalizeBashRunError(context.Background(), exec.ErrWaitDelay, true); err != nil {
		t.Fatalf("preserved post-exit WaitDelay should be ignored, got %v", err)
	}
	if err := normalizeBashRunError(context.Background(), exec.ErrWaitDelay, false); !errors.Is(err, exec.ErrWaitDelay) {
		t.Fatalf("ordinary WaitDelay should remain visible, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := normalizeBashRunError(ctx, exec.ErrWaitDelay, true); !errors.Is(err, exec.ErrWaitDelay) {
		t.Fatalf("cancelled WaitDelay should remain visible, got %v", err)
	}
}
