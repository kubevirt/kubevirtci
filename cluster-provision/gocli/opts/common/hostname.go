package common

import "fmt"

// WaitForHostnameCommand waits for a successful hostname command to report exactly
// expectedHostname (an internal node name such as node01 or node02). It checks
// immediately and after each of six 5-second sleeps, including at 30 seconds.
// The POSIX shell command returns zero on readiness and nonzero on timeout or
// sleep failure, reporting waits to stdout and timeout details to stderr.
func WaitForHostnameCommand(expectedHostname string) string {
	return fmt.Sprintf(`(
expected_hostname=%q
elapsed=0
while :; do
    if observed_hostname=$(hostname); then
        hostname_status=0
    else
        hostname_status=$?
    fi
    if [ "$hostname_status" -eq 0 ] && [ "$observed_hostname" = "$expected_hostname" ]; then
        break
    fi
    if [ "$elapsed" -ge 30 ]; then
        printf 'Timed out after 30 seconds waiting for hostname: expected "%%s", observed "%%s" (hostname status %%s)\n' "$expected_hostname" "$observed_hostname" "$hostname_status" >&2
        exit 1
    fi
    printf 'Waiting for hostname: expected "%%s", observed "%%s" (hostname status %%s); retrying in 5 seconds\n' "$expected_hostname" "$observed_hostname" "$hostname_status"
    sleep 5 || exit $?
    elapsed=$((elapsed + 5))
done
)`, expectedHostname)
}
