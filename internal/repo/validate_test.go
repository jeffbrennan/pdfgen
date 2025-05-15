package repo

import (
	"testing"

	"github.com/jeffbrennan/pdfgen/internal/models"
)

func TestValidateRepo(t *testing.T) {
	var tests = []struct {
		name       string
		input      *models.RepoParts
		shouldPass bool
	}{
		{
			"airflow should pass",
			&models.RepoParts{
				Provider: "github.com",
				Owner:    "apache",
				Repo:     "airflow"},
			true},
		{
			"ampere should fail",
			&models.RepoParts{
				Provider: "github.com",
				Owner:    "jeffbrennan",
				Repo:     "ampere"},
			false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepo(tt.input)
			if (err != nil) && (tt.shouldPass) {
				t.Errorf("should pass: got %v", err)
			}

			if (err == nil) && !(tt.shouldPass) {
				t.Errorf("should fail")
			}
		})
	}
}
