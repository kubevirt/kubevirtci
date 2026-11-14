package common

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWaitForHostnameCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hostname Readiness Suite")
}

type hostnameTestCase struct {
	expected       string
	observed       string
	hostnameStatus int
	readyAt        int
	sleepFails     bool
	wantSleeps     int
	wantStatus     int
}

var _ = Describe("Hostname readiness command", func() {
	DescribeTable("waits for the exact expected hostname",
		func(tt hostnameTestCase) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			DeferCleanup(cancel)

			// Only sleep mutates attempt: hostname runs in a command substitution.
			stubs := fmt.Sprintf(`attempt=0
hostname() {
    if [ %d -ge 0 ] && [ "$attempt" -ge %d ]; then
        printf '%%s\n' %q
        return 0
    fi
    printf '%%s\n' %q
    return %d
}
sleep() {
    [ "$#" -eq 1 ] && [ "$1" = 5 ] || return 99
    printf 'sleep %%s\n' "$1"
    if [ %t = true ]; then
        printf 'stub sleep failed\n' >&2
        return 7
    fi
    attempt=$((attempt + 1))
}
`, tt.readyAt, tt.readyAt, tt.expected, tt.observed, tt.hostnameStatus, tt.sleepFails)
			cmd := exec.CommandContext(ctx, "sh", "-eu", "-c", stubs+WaitForHostnameCommand(tt.expected))
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			Expect(ctx.Err()).NotTo(HaveOccurred(), "shell exceeded context timeout")
			status := 0
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				Expect(ok).To(BeTrue(), "execute shell: %v", err)
				status = exitErr.ExitCode()
			}
			Expect(status).To(Equal(tt.wantStatus), "stdout: %q; stderr: %q", stdout.String(), stderr.String())

			wait := fmt.Sprintf("Waiting for hostname: expected %q, observed %q (hostname status %d); retrying in 5 seconds\n", tt.expected, tt.observed, tt.hostnameStatus)
			wantStdout := strings.Repeat(wait+"sleep 5\n", tt.wantSleeps)
			Expect(stdout.String()).To(Equal(wantStdout))
			wantStderr := ""
			if tt.sleepFails {
				wantStderr = "stub sleep failed\n"
			} else if tt.wantStatus != 0 {
				wantStderr = fmt.Sprintf("Timed out after 30 seconds waiting for hostname: expected %q, observed %q (hostname status %d)\n", tt.expected, tt.observed, tt.hostnameStatus)
			}
			Expect(stderr.String()).To(Equal(wantStderr))
		},
		Entry("node01 immediately ready", hostnameTestCase{expected: "node01", readyAt: 0}),
		Entry("node02 immediately ready", hostnameTestCase{expected: "node02", readyAt: 0}),
		Entry("localhost then desired", hostnameTestCase{expected: "node01", observed: "localhost", readyAt: 2, wantSleeps: 2}),
		Entry("different node is not ready", hostnameTestCase{expected: "node02", observed: "node03", readyAt: -1, wantSleeps: 6, wantStatus: 1}),
		Entry("empty output is not ready", hostnameTestCase{expected: "node01", readyAt: -1, wantSleeps: 6, wantStatus: 1}),
		Entry("matching output with command failure", hostnameTestCase{expected: "node02", observed: "node02", hostnameStatus: 3, readyAt: -1, wantSleeps: 6, wantStatus: 1}),
		Entry("command failure then recovery", hostnameTestCase{expected: "node01", observed: "node01", hostnameStatus: 3, readyAt: 1, wantSleeps: 1}),
		Entry("never ready", hostnameTestCase{expected: "node01", observed: "localhost", readyAt: -1, wantSleeps: 6, wantStatus: 1}),
		Entry("ready on final check", hostnameTestCase{expected: "node02", observed: "localhost", readyAt: 6, wantSleeps: 6}),
		Entry("sleep failure stops retries", hostnameTestCase{expected: "node01", observed: "localhost", readyAt: 1, sleepFails: true, wantSleeps: 1, wantStatus: 7}),
	)
})
