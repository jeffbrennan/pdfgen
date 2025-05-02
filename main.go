package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type RepoParts struct {
	provider  string
	owner     string
	repo      string
	branch    string
	directory string
}

func parseRepoURL(url string) (*RepoParts, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("invalid URL: %s", url)
	}

	if !strings.Contains(url, "github.com") {
		return nil, fmt.Errorf("unsupported provider: %s", url)
	}

	url = strings.TrimSuffix(strings.TrimPrefix(url, "https://"), "/")
	url = strings.Replace(url, "/tree/", "/", 1)

	// variants
	// github.com/apache/airflow/tree/main/airflow-core/docs # 6
	// github.com/apache/airflow #2
	// github.com/apache/airflow/airflow-core/docs # 4
	parts := strings.Split(url, "/")
	nSlashes := len(parts)

	// need at least provder, owner, repo
	if nSlashes < 3 {
		return nil, fmt.Errorf("invalid URL: %s", url)
	}

	provider := parts[0]
	owner := parts[1]
	repo := parts[2]
	if nSlashes == 3 {
		return &RepoParts{
			provider: provider,
			owner:    owner,
			repo:     repo,
		}, nil
	}

	branch := ""
	branchPadding := 0
	if parts[4] == "tree" {
		branch = parts[5]
		branchPadding = 2
	} else {
		branch = "main"
	}

	if nSlashes == 4+branchPadding {
		return &RepoParts{
			provider:  provider,
			owner:     owner,
			repo:      repo,
			branch:    branch,
			directory: "docs",
		}, nil

	}

	directory := strings.Join(parts[4+branchPadding:], "/")

	return &RepoParts{
		provider:  provider,
		owner:     owner,
		repo:      repo,
		branch:    branch,
		directory: directory,
	}, nil

}

func cloneRepo(parts *RepoParts) error {
	log.Printf("Cloning %s/%s/%s", parts.provider, parts.owner, parts.repo)

	log.Printf("Creating directory /tmp")
	err := exec.Command("mkdir", "-p", "/tmp/pdfgen").Run()
	if err != nil {
		return err
	}

	baseURL := fmt.Sprintf(
		"https://%s/%s/%s.git",
		parts.provider,
		parts.owner,
		parts.repo,
	)
	cmd := exec.Command(
		"git",
		"clone",
		baseURL,
		"-b",
		parts.branch,
		"--single-branch",
		"--depth",
		"1",
	)
	cmd.Dir = "/tmp/pdfgen"

	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// receive url from client
	url := "https://github.com/apache/airflow/tree/main/airflow-core/docs"
	parsedURL, err := parseRepoURL(url)
	if err != nil {
		log.Fatal("Error parsing URL:", err)
	}

	fmt.Printf("%+v\n", parsedURL)
	err = cloneRepo(parsedURL)
	if err != nil {
		log.Fatal("Error cloning repo:", err)
	}

	// navigate to optional directory
	// identify documenation format
	// run tool to generate pdf
	// return pdf to client
}
