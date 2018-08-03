package processor

import (
	"testing"

	"github.com/gocraft/work"
)

func TestContext_CheckTransaction(t *testing.T) {
	type args struct {
		job *work.Job
	}
	tests := []struct {
		name    string
		c       *Context
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{}
			if err := c.CheckTransaction(tt.args.job); (err != nil) != tt.wantErr {
				t.Errorf("Context.CheckTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
