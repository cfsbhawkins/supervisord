package process

import (
	"testing"
	"time"
)

func waitState(t *testing.T, p *Process, want State, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for p.GetState() != want {
		if time.Now().After(deadline) {
			t.Fatalf("%s: state %v, want %v", p.GetName(), p.GetState(), want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Stop(true) immediately followed by Start(true) is what the liveness-probe
// "restart" action and XML-RPC clients (stopProcess, then startProcess) do.
// Start used to be dropped while the previous run loop was still winding
// down; for a program that had run under two seconds that loop first slept
// five seconds, so the program stayed STOPPED and StartProcess answered
// "Fail to start process".
func TestStartRightAfterStopRestartsProgram(t *testing.T) {
	for _, autorestart := range []string{"true", "false"} {
		t.Run("autorestart="+autorestart, func(t *testing.T) {
			p := newTestProcess(t, "quick", `[program:quick]
command=/bin/sh -c "exec sleep 30"
startsecs=0
autorestart=`+autorestart+`
stopwaitsecs=5
`)
			p.Start(true)
			waitState(t, p, Running, 5*time.Second)
			p.Stop(true)
			waitState(t, p, Stopped, 5*time.Second)

			begin := time.Now()
			p.Start(true)
			waitState(t, p, Running, 5*time.Second)
			if d := time.Since(begin); d > 3*time.Second {
				t.Fatalf("restart took %v; the run loop's 5s pause must not apply after a user stop", d)
			}

			p.Stop(true)
			waitState(t, p, Stopped, 5*time.Second)
		})
	}
}

// Start on a program that is already running stays a no-op: the same
// process keeps running and no second instance is spawned.
func TestStartWhileRunningIsNoop(t *testing.T) {
	p := newTestProcess(t, "steady", `[program:steady]
command=/bin/sh -c "exec sleep 30"
startsecs=0
autorestart=false
stopwaitsecs=5
`)
	p.Start(true)
	waitState(t, p, Running, 5*time.Second)
	pid := p.GetPid()

	p.Start(true)
	if p.GetState() != Running {
		t.Fatalf("state %v after second Start, want Running", p.GetState())
	}
	if got := p.GetPid(); got != pid {
		t.Fatalf("pid changed from %d to %d: program was started twice", pid, got)
	}

	p.Stop(true)
	waitState(t, p, Stopped, 5*time.Second)
}
