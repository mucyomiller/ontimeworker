package enqueue

import "testing"

func TestEnqueue(t *testing.T) {
	type args struct {
		tx map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Enqueue(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enqueue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Enqueue() = %v, want %v", got, tt.want)
			}
		})
	}
}
