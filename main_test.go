package main

import (
	"net/http"
	"testing"

	_ "github.com/joho/godotenv/autoload"
)

func Test_handleJob(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleJob(tt.args.w, tt.args.r)
		})
	}
}
