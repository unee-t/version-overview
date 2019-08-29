package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_versionString(t *testing.T) {
	ae5b321, err := os.Open("testdata/ae5b321.txt")
	if err != nil {
		t.Fatal(err)
	}
	foo, err := os.Open("testdata/foo.txt")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		input io.Reader
	}
	tests := []struct {
		name        string
		args        args
		wantVersion string
		wantErr     bool
	}{
		{
			name: "ae5b321",
			args: args{
				input: bufio.NewReader(ae5b321),
			},
			wantVersion: "a83551366ab74bf43ce8c6019b94c5329d81eaf1",
			wantErr:     false,
		},
		{
			name: "foo",
			args: args{
				input: bufio.NewReader(foo),
			},
			wantVersion: "foo",
			wantErr:     false,
		},
		{
			name: "edge",
			args: args{
				input: strings.NewReader(`<!-- --> <!-- COMMIT: 8f0c86e5cdb1fe342912b4975556eb86a6536234 -->`),
			},
			wantVersion: "8f0c86e5cdb1fe342912b4975556eb86a6536234",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, err := parseVersion(tt.args.input, "<!-- COMMIT: ", " -->")
			if (err != nil) != tt.wantErr {
				t.Errorf("versionString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("versionString() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}
