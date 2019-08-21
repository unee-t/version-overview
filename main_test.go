package main

import (
	"bufio"
	"io"
	"os"
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
			wantVersion: "ae5b321",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, err := commitVersion(tt.args.input)
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
