package main

import "testing"

func TestValidateRepo(t *testing.T) {
	var tests = []struct {
		name       string
		input      *RepoParts
		shouldPass bool
	}{
		{"airflow should pass", &RepoParts{provider: "github.com", owner: "apache", repo: "airflow"}, true},
		{"ampere should fail", &RepoParts{provider: "github.com", owner: "jeffbrennan", repo: "ampere"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepo(tt.input)
			if (err != nil) && (tt.shouldPass) {
				t.Errorf("should pass: got %v", err)
			}

			if (err == nil) && !(tt.shouldPass) {
				t.Errorf("should fail")
			}
		})
	}
}
