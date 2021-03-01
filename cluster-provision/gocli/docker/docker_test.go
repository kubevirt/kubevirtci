package docker

import (
	"reflect"
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
					container("prefix-1", "something", "unimportant"),
					container("prefix-2", "something1", "unimportant1"),
					container("absolutely-unrelated"),
					container("something1", "not-a-prefix1", "unimportant1"),
					container("something1", "prefix-3", "unimportant1"),
					container("prefix-4"),
					container("/prefix-5"),
					container("not-a-prefix2", "unimportant1"),
				},
				prefix: "prefix",
			},
			want: []types.Container{
				container("prefix-1", "something", "unimportant"),
				container("prefix-2", "something1", "unimportant1"),
				container("something1", "prefix-3", "unimportant1"),
				container("prefix-4"),
				container("/prefix-5"),
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

func container(names ...string) types.Container {
	return types.Container{
		Names: names,
	}
}
