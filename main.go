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

type DirectoryParts struct {
	base string
	doc  string
}

type DocumentationFormat int

const (
	Sphinx DocumentationFormat = iota
	MkDocs
	Docusaurus
	GitBook
)

var documentationName = map[DocumentationFormat]string{
	Sphinx:     "sphinx",
	MkDocs:     "mkdocs",
	Docusaurus: "docusaurus",
	GitBook:    "gitbook",
}

func generatePDF(dirParts *DirectoryParts, docType DocumentationFormat) error {
	var cmd []string

	// TODO: investigate if relative doc path is needed instead
	// TODO: sphinx requires python env to match cloned env
	// need to support requirements.txt, poetry, uv
	switch docType {
	case Sphinx:
		cmd = []string{
			"sphinx-build",
			"-M",
			"simplepdf",
			dirParts.doc,
		}
	case MkDocs:
		cmd = []string{"mkdocs", "build"}
	case Docusaurus:
		cmd = []string{"npm", "run", "build"}
	case GitBook:
		cmd = []string{"gitbook", "build"}
	}

	_, err := RunCommand(cmd, dirParts.doc)
	return err

}

func parseDocumentationFormat(
	dirParts *DirectoryParts,
) (DocumentationFormat, error) {
	out, err := RunCommand([]string{"ls", dirParts.doc}, dirParts.base)
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "conf.py") ||
			strings.HasSuffix(file, "index.rst") {
			log.Printf("Found Sphinx documentation: %s", file)
			return Sphinx, nil
		}

		if strings.HasSuffix(file, "mkdocs.yml") ||
			strings.HasSuffix(file, "mkdocs.yaml") {
			log.Printf("Found MkDocs documentation: %s", file)
			return MkDocs, nil
		}
		if strings.HasSuffix(file, "docusaurus.config.js") {
			log.Printf("Found Docusaurus documentation: %s", file)
			return Docusaurus, nil
		}
		if strings.HasSuffix(file, "gitbook.yml") ||
			strings.HasSuffix(file, "gitbook.yaml") {
			log.Printf("Found GitBook documentation: %s", file)
			return GitBook, nil
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

func RunCommand(args []string, workingDir string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	log.Printf("Executing: %s", strings.Join(cmd.Args, " "))
	return cmd.Output()
}

func pullRepo(targetDir string) error {
	_, err := RunCommand([]string{"git", "pull"}, targetDir)
	return err
}

func parseRepoDir(parts *RepoParts) (*DirectoryParts, error) {
	baseDir := "/tmp/pdfgen"
	targetDir := fmt.Sprintf(
		"%s/%s", baseDir, parts.repo,
	)
	var docDir string
	if parts.directory == "" {
		docDir = fmt.Sprintf("%s/docs", targetDir)
	} else {
		docDir = fmt.Sprintf("%s/%s", targetDir, parts.directory)
	}

	dirParts := &DirectoryParts{
		base: targetDir,
		doc:  docDir,
	}

	return dirParts, nil
}

func updateRepo(parts *RepoParts) error {
	// pulls or clones depending on if the repo exists
	baseDir := "/tmp/pdfgen"
	targetDir := fmt.Sprintf(
		"%s/%s", baseDir, parts.repo,
	)

	log.Printf("Updating %s/%s/%s", parts.provider, parts.owner, parts.repo)
	_, err := RunCommand([]string{"ls"}, targetDir)
	if err == nil {
		log.Printf("Directory %s already exists", targetDir)
		err = pullRepo(targetDir)
		if err != nil {
			return err
		}
		log.Printf("Pulled latest changes for %s", targetDir)
		return nil
	}

	_, err = RunCommand([]string{"mkdir", "-p", baseDir}, "")
	if err != nil {
		return err
	}

	return cloneRepo(parts, baseDir)
}

func cloneRepo(parts *RepoParts, baseDir string) error {
	baseURL := fmt.Sprintf(
		"https://%s/%s/%s.git",
		parts.provider,
		parts.owner,
		parts.repo,
	)
	_, err := RunCommand(
		[]string{
			"git",
			"clone",
			baseURL,
			"-b",
			parts.branch,
			"--single-branch",
			"--depth",
			"1",
		},
		baseDir,
	)
	return err
}

func main() {
	// receive url from client
	url := "https://github.com/apache/airflow/tree/main/airflow-core/docs"
	parsedURL, err := parseRepoURL(url)
	if err != nil {
		log.Fatal("Error parsing URL:", err)
	}

	err = updateRepo(parsedURL)
	if err != nil {
		log.Fatal("Error updating repo:", err)
	}

	dirParts, err := parseRepoDir(parsedURL)
	if err != nil {
		log.Fatal("Error parsing repo directory:", err)
	}

	docName, err := parseDocumentationFormat(dirParts)
	if err != nil {
		log.Fatal("Error parsing documentation format:", err)
	}

	fmt.Printf("Documentation format: %s\n", documentationName[docName])
	err = generatePDF(dirParts, docName)
	if err != nil {
		log.Fatal("Error generating PDF:", err)
	}

	// return pdf to client

	// cleanup
}
