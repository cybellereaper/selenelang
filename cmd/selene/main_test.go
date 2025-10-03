package main

import "testing"

func TestValidateLSPArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: nil, wantErr: false},
		{name: "stdio flag", args: []string{"--stdio"}, wantErr: false},
		{name: "stdio true", args: []string{"--stdio=true"}, wantErr: false},
		{name: "stdio false", args: []string{"--stdio=false"}, wantErr: true},
		{name: "positional", args: []string{"extra"}, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := validateLSPArgs(tt.args)
			if tt.wantErr && err == nil {
				t.Fatalf("validateLSPArgs(%v) = nil error, want error", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateLSPArgs(%v) unexpected error: %v", tt.args, err)
			}
		})
	}
}
