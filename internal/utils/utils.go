package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jeffbrennan/pdfgen/internal/models"
)

func LoadSecret(k string) (string, error) {
	secretPath := fmt.Sprintf("/run/secrets/%s", k)
	data, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatalf("failed to read secret from %s: %v", secretPath, err)
	}

	secretValue := strings.TrimSpace(strings.Split(string(data), "=")[1])
	return secretValue, nil
}

func RunCommand(args []string, workingDir string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	return cmd.Output()
}

func CleanupDir(parts *models.RepoParts) error {
	repoPath := "/build/src/pdfgen/repos/" + parts.Repo
	log.Printf("Cleaning up %s/%s/%s", parts.Provider, parts.Owner, parts.Repo)
	_, err := RunCommand([]string{"rm", "-rf", repoPath}, "")
	return err
}
