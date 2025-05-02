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

type DocumentationFormat int

const (
	Sphinx DocumentationFormat = iota
	MkDocs
	Doxygen
)

var documentationName = map[DocumentationFormat]string{
	Sphinx:  "sphinx",
	MkDocs:  "mkdocs",
	Doxygen: "doxygen",
}

func parseDocumentationFormat(
	cloneDir string,
	parts *RepoParts,
) (DocumentationFormat, error) {
	var docDir string

	if parts.directory == "" {
		docDir = fmt.Sprintf("%s/docs", cloneDir)
	} else {
		docDir = fmt.Sprintf("%s/%s", cloneDir, parts.directory)
	}

	cmd := exec.Command("ls", docDir)
	cmd.Dir = cloneDir
	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	out, err := cmd.Output()
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "conf.py") {
			log.Printf("Found Sphinx documentation: %s", file)
			return Sphinx, nil
		}

		if strings.HasSuffix(file, "index.rst") {
			log.Printf("Found Sphinx documentation: %s", file)
			return Sphinx, nil
		}

		if strings.HasSuffix(file, "mkdocs.yml") {
			log.Printf("Found MkDocs documentation: %s", file)
			return MkDocs, nil
		}
		if strings.HasSuffix(file, "Doxyfile") {
			log.Printf("Found Doxygen documentation: %s", file)
			return Doxygen, nil
		}
	}

	return -1, fmt.Errorf("unknown documentation format")

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

func pullRepo(targetDir string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = targetDir
	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func cloneRepo(parts *RepoParts) (string, error) {
	baseDir := "/tmp/pdfgen"
	targetDir := fmt.Sprintf(
		"%s/%s", baseDir, parts.repo,
	)

	log.Printf("Cloning %s/%s/%s", parts.provider, parts.owner, parts.repo)

	cmd := exec.Command("ls", targetDir)
	err := cmd.Run()
	if err == nil {
		log.Printf("Directory %s already exists", targetDir)
		err = pullRepo(targetDir)
		if err != nil {
			return "", err
		}
		log.Printf("Pulled latest changes for %s", targetDir)
		return targetDir, nil

	}

	log.Printf("Creating directory %s", baseDir)
	err = exec.Command("mkdir", "-p", baseDir).Run()
	if err != nil {
		return "", err
	}

	baseURL := fmt.Sprintf(
		"https://%s/%s/%s.git",
		parts.provider,
		parts.owner,
		parts.repo,
	)
	cmd = exec.Command(
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
		return "", err
	}

	return targetDir, nil
}

func main() {
	// receive url from client
	url := "https://github.com/apache/airflow/tree/main/airflow-core/docs"
	parsedURL, err := parseRepoURL(url)
	if err != nil {
		log.Fatal("Error parsing URL:", err)
	}

	fmt.Printf("%+v\n", parsedURL)
	cloneDir, err := cloneRepo(parsedURL)
	if err != nil {
		log.Fatal("Error cloning repo:", err)
	}

	docName, err := parseDocumentationFormat(cloneDir, parsedURL)
	if err != nil {
		log.Fatal("Error parsing documentation format:", err)
	}

	fmt.Printf("Documentation format: %s\n", documentationName[docName])
	// run tool to generate pdf
	// return pdf to client
}
