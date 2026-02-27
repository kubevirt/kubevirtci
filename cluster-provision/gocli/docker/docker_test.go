package docker

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
)

func Test_filterByPrefix(t *testing.T) {
	type args struct {
		containers []types.Container
		prefix     string
	}
	tests := []struct {
		name string
		args args
		want []types.Container
	}{
		{name: "should filter by prefix",
			args: struct {
				containers []types.Container
				prefix     string
			}{
				containers: []types.Container{
					containerFromNames("prefix-1", "something", "unimportant"),
					containerFromNames("prefix-2", "something1", "unimportant1"),
					containerFromNames("absolutely-unrelated"),
					containerFromNames("something1", "not-a-prefix1", "unimportant1"),
					containerFromNames("something1", "prefix-3", "unimportant1"),
					containerFromNames("prefix-4"),
					containerFromNames("/prefix-5"),
					containerFromNames("not-a-prefix2", "unimportant1"),
				},
				prefix: "prefix",
			},
			want: []types.Container{
				containerFromNames("prefix-1", "something", "unimportant"),
				containerFromNames("prefix-2", "something1", "unimportant1"),
				containerFromNames("something1", "prefix-3", "unimportant1"),
				containerFromNames("prefix-4"),
				containerFromNames("/prefix-5"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterByPrefix(tt.args.containers, tt.args.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterByPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func containerFromNames(names ...string) types.Container {
	return types.Container{
		Names: names,
	}
}

func TestPrintProgress_NonTerminal_ShowsStatusAndPercent(t *testing.T) {
	// Fake JSON stream similar to Docker pull output
	jsonStream := `
{"status":"Downloading","id":"sha256:abc","progressDetail":{"current":50,"total":100}}
{"status":"Download complete","id":"sha256:abc"}
`

	// Reader simulates what cli.ImagePull returns
	r := io.NopCloser(strings.NewReader(jsonStream))

	// Writer is a temporary file (not a terminal), so we hit the non-terminal branch
	f, err := os.CreateTemp("", "print-progress-test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if err := PrintProgress(r, f); err != nil {
		t.Fatalf("PrintProgress returned error: %v", err)
	}

	// Read back what was written
	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	out := strings.TrimSpace(string(data))
	if out == "" {
		t.Fatalf("expected some output, got none")
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		t.Fatalf("expected at least one line of output, got: %q", out)
	}

	// Find a line that contains "Downloading"
	var downloadingLine string
	for _, line := range lines {
		if strings.Contains(line, "Downloading") {
			downloadingLine = line
			break
		}
	}

	if downloadingLine == "" {
		t.Fatalf("expected at least one line containing 'Downloading', got: %q", out)
	}

	// We expect the downloading line to include:
	// - the status "Downloading"
	// - the layer id "sha256:abc"
	// - the computed percent "50%"
	if !strings.Contains(downloadingLine, "sha256:abc") {
		t.Fatalf("expected downloading line to contain 'sha256:abc', got: %q", downloadingLine)
	}
	if !strings.Contains(downloadingLine, "50%") {
		t.Fatalf("expected downloading line to contain '50%%', got: %q", downloadingLine)
	}

}
