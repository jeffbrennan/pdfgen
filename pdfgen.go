package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type RepoParts struct {
	provider  string
	owner     string
	repo      string
	branch    string
	directory string
}

type DirectoryParts struct {
	root string
	base string
	doc  string
}

type GithubRepoResponse struct {
	StargazersCount int       `json:"stargazers_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type RepoStats struct {
	stars    int
	ageYears float64
}

type PDFGenResponse struct {
	parts    *RepoParts
	dirParts *DirectoryParts
	pdfPath  string
	pdfBytes []byte
}

type DocumentationFormat int
type PythonEnv int
type EnvType int

const (
	Sphinx DocumentationFormat = iota
	MkDocs
	Docusaurus
	GitBook
)

const (
	pip PythonEnv = iota
	poetry
	uv
)

const (
	python EnvType = iota
	node
)

var documentationName = map[DocumentationFormat]string{
	Sphinx:     "sphinx",
	MkDocs:     "mkdocs",
	Docusaurus: "docusaurus",
	GitBook:    "gitbook",
}

func setupPythonEnvPip(dirParts *DirectoryParts) error {

	_, err := RunCommand(
		[]string{"uv", "pip", "install", "-r", "requirements.txt"},
		dirParts.base,
	)
	return err

}

func setupPythonEnvPoetry(dirParts *DirectoryParts) error {
	// TODO: parse pyproject.toml to look for a docs group
	_, err := RunCommand(
		[]string{"uvx", "migrate-to-uv"},
		dirParts.base,
	)

	if err != nil {
		return err
	}

	_, err = RunCommand(
		[]string{"uv", "sync"},
		dirParts.base,
	)

	return err
}

func setupPythonEnvUV(dirParts *DirectoryParts) error {
	_, err := RunCommand(
		[]string{"uv", "sync"},
		dirParts.base,
	)

	return err
}

func setupPythonEnv(dirParts *DirectoryParts, env PythonEnv) error {
	publishLog("Setting up Python environment...")
	_, err := RunCommand(
		[]string{"uv", "venv"},
		dirParts.base,
	)

	if err != nil {
		return err
	}

	switch env {
	case pip:
		return setupPythonEnvPip(dirParts)
	case poetry:
		return setupPythonEnvPoetry(dirParts)
	case uv:
		return setupPythonEnvUV(dirParts)
	default:
		return fmt.Errorf("unknown python env")
	}
}

func setupNodeEnv(dirParts *DirectoryParts) error {
	return fmt.Errorf("node env setup not implemented")
}

func handleSphinxIssuesVersionKeyError(dirParts *DirectoryParts) error {
	// workaround for airflow build - should generalize after testing other sphinx builds
	extDir := "devel-common/src/sphinx_exts/"
	out, err := RunCommand([]string{"ls", extDir}, dirParts.root)
	log.Print(out)
	if err != nil {
		log.Print("Error running ls: ", err)
		return err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if !strings.Contains(file, "substitution_extensions.py") {
			continue
		}

		filePath := dirParts.root + "/" + extDir + file
		fileHandle, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fileHandle.Close()

		contentBytes, err := io.ReadAll(fileHandle)
		if err != nil {
			log.Printf("Error reading file: %s", err)
			return err
		}

		log.Print("writing file...")
		newFileContents := strings.ReplaceAll(
			string(contentBytes),
			"version = substitution_defs[\"version\"].astext()",
			"version = substitution_defs.get(\"version\", \"unknown\")",
		)
		err = os.WriteFile(filePath, []byte(newFileContents), 0644)
		if err != nil {
			print("Error writing file: %s", err)
			return err
		}

	}

	return nil
}

func handleSphinxIssues(dirParts *DirectoryParts) error {
	err := handleSphinxIssuesVersionKeyError(dirParts)
	return err
}

func generateSphinxPDF(parts *RepoParts, dirParts *DirectoryParts) (string, error) {
	err := handleSphinxIssues(dirParts)
	if err != nil {
		log.Printf("Error handling Sphinx issues: %s", err)
		return "", err
	}

	// TODO: handle case where docs group does not exist
	publishLog("Generating docs as Latex...")
	out, err := RunCommand([]string{
		"uv",
		"run",
		"--group",
		"docs",
		"sphinx-build",
		"-M",
		"latex",
		dirParts.doc,
		"_build/",
	},
		dirParts.base,
	)

	log.Printf("Sphinx build output: %s\n", out)
	if err != nil {
		log.Printf("Error running sphinx-build: %s", out)
		log.Printf("Error: %s", err)
		return "", err
	}

	outputName := parts.repo + "_" + strings.ReplaceAll(parts.directory, "/", "_")

	publishLog("Converting Latex to PDF...")
	out, err = RunCommand(
		[]string{
			"/bin/sh",
			"-c",
			"pdflatex -interaction=nonstopmode -jobname=" + outputName + " $(find -maxdepth 1 -name '*.tex' | head -n 1)",
		},
		dirParts.base+"/_build/latex",
	)
	log.Printf("finished running pdflatex in %s\n", dirParts.base+"/_build/latex")

	// ignore for now
	log.Printf("uncaught pdflatex output: %s\n", out)
	log.Printf("uncaught pdflatex error: %s\n", err)

	// if err != nil {
	// 	log.Printf("Error running pdflatex: %s", out)
	// 	log.Printf("Error: %s", err)
	// 	return "", err
	// }

	pdfPath := dirParts.base + "/_build/latex/" + outputName + ".pdf"
	log.Printf("PDF path: %s", pdfPath)
	return pdfPath, nil
}

func generateMkDocsPDF(dirParts *DirectoryParts) (string, error) {
	return "", fmt.Errorf("MkDocs PDF generation not implemented")
}

func generateDocusaurusPDF(dirParts *DirectoryParts) (string, error) {
	return "", fmt.Errorf("docusaurus PDF generation not implemented")
}

func generateGitBookPDF(dirParts *DirectoryParts) (string, error) {
	return "", fmt.Errorf("GitBook PDF generation not implemented")
}

func generatePDF(parts *RepoParts, dirParts *DirectoryParts, docType DocumentationFormat) (string, error) {
	publishLog("Generating PDF...")
	envType, err := parseEnvType(dirParts)
	if err != nil {
		return "", err
	}
	if envType == python {
		pythonEnv, err := parsePythonEnv(dirParts)
		if err != nil {
			return "", err
		}

		setupPythonEnv(dirParts, pythonEnv)
	}

	switch docType {
	case Sphinx:
		return generateSphinxPDF(parts, dirParts)
	case MkDocs:
		return generateMkDocsPDF(dirParts)
	case Docusaurus:
		return generateDocusaurusPDF(dirParts)
	case GitBook:
		return generateGitBookPDF(dirParts)
	}

	return "", fmt.Errorf("unknown documentation format")

}

func parseEnvType(dirParts *DirectoryParts) (EnvType, error) {
	out, err := RunCommand([]string{"ls"}, dirParts.base)
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "package.json") {
			log.Printf("Found node env: %s", file)
			return node, nil
		}
		if strings.HasSuffix(file, "requirements.txt") ||
			strings.HasSuffix(file, "poetry.lock") ||
			strings.HasSuffix(file, "uv.lock") ||
			strings.HasSuffix(file, "pyproject.toml") {
			log.Printf("Found python env: %s", file)
			return python, nil
		}
	}

	return -1, fmt.Errorf("unknown env")
}

func parsePythonEnv(dirParts *DirectoryParts) (PythonEnv, error) {
	publishLog("Parsing Python env...")
	out, err := RunCommand([]string{"ls"}, dirParts.base)
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "requirements.txt") {
			log.Printf("Found pip env: %s", file)
			return pip, nil
		}

		if strings.HasSuffix(file, "uv.lock") {
			log.Printf("Found uv: %s", file)
			return uv, nil
		}
		if strings.HasSuffix(file, "poetry.lock") || strings.HasSuffix(file, "pyproject.toml") {
			return poetry, nil
		}
	}

	return -1, fmt.Errorf("unknown env")

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
			sphinxMsg := fmt.Sprintf(
				"Found Sphinx documentation: %s",
				file,
			)
			log.Print(sphinxMsg)
			publishLog(sphinxMsg)

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
	var docDir string
	var baseDir string

	rootDir := fmt.Sprintf(
		"%s/%s", "./repos", parts.repo,
	)

	if parts.directory == "" {
		baseDir = rootDir
		docDir = "docs/"
	} else {

		docParts := strings.Split(parts.directory, "/")
		// strip last element
		baseDir = rootDir + "/" + strings.Join(docParts[:len(docParts)-1], "/")
		log.Printf("Base dir: %s", baseDir)

		docDir = docParts[len(docParts)-1] + "/"
	}

	dirParts := &DirectoryParts{
		root: rootDir, //  .../airflow
		base: baseDir, // .../airflow/airflow-core/
		doc:  docDir,  // docs/
	}

	return dirParts, nil
}

func updateRepo(parts *RepoParts) error {
	// pulls or clones depending on if the repo exists
	baseDir := "./repos"
	targetDir := fmt.Sprintf(
		"%s/%s", baseDir, parts.repo,
	)

	updateRepoMsg := fmt.Sprintf(
		"Updating %s/%s/%s...",
		parts.provider,
		parts.owner,
		parts.repo,
	)

	log.Print(updateRepoMsg)
	publishLog(updateRepoMsg)

	_, err := RunCommand([]string{"ls"}, targetDir)
	if err == nil {
		log.Printf("Directory %s already exists", targetDir)
		err = pullRepo(targetDir)
		if err != nil {
			return err
		}
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

func cleanupRepo(parts *RepoParts) error {
	repoPath := "/build/src/pdfgen/repos/" + parts.repo
	log.Printf("Cleaning up %s/%s/%s", parts.provider, parts.owner, parts.repo)
	_, err := RunCommand([]string{"rm", "-rf", repoPath}, "")
	return err
}

func getGithubAPIResponseRepo(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%v/%v", owner, repo)
	log.Printf("requesting a response from %s...", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN not set in .env file")
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bodyText, nil
}

func parseGithubAPIResponse(responseBody []byte) (*RepoStats, error) {
	var response GithubRepoResponse
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return &RepoStats{}, err
	}
	ageYears := time.Since(response.CreatedAt).Hours() / (24 * 365.25)

	return &RepoStats{
		stars:    response.StargazersCount,
		ageYears: ageYears,
	}, nil

}

func validateRepo(parts *RepoParts) error {
	// guard against improper usage only accepting large, established projects
	minNumStars := 100
	minNumStarsNewRepo := 1000
	minRepoAgeYears := 1.0

	response, err := getGithubAPIResponseRepo(parts.owner, parts.repo)
	if err != nil {
		return err
	}

	repoStats, err := parseGithubAPIResponse(response)
	if err != nil {
		return err
	}

	if repoStats.stars < minNumStars {
		return fmt.Errorf("repo has less than %d stars", minNumStars)
	}

	if (repoStats.ageYears < minRepoAgeYears) && (repoStats.stars < minNumStarsNewRepo) {
		return fmt.Errorf("repo is less than %f years old and has less than %d stars", minRepoAgeYears, minNumStarsNewRepo)
	}

	log.Printf("%s/%s is valid", parts.owner, parts.repo)
	return nil
}

func pdfgen(url string) (PDFGenResponse, error) {
	parts, err := parseRepoURL(url)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error parsing URL: %s", err)
	}

	err = validateRepo(parts)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error validating repo: %s", err)
	}

	err = updateRepo(parts)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error updating repo: %s", err)
	}

	dirParts, err := parseRepoDir(parts)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error parsing repo directory: %s", err)
	}

	docName, err := parseDocumentationFormat(dirParts)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error parsing documentation format: %s", err)
	}

	log.Printf("Documentation format: %s\n", documentationName[docName])
	pdfPath, err := generatePDF(parts, dirParts, docName)
	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error generating PDF: %s", err)
	}

	log.Printf("reading PDF file: %s", pdfPath)
	pdfBytes, err := os.ReadFile(pdfPath)

	if err != nil {
		return PDFGenResponse{}, fmt.Errorf("error reading PDF file: %s", err)
	}
	publishLog("done!")

	return PDFGenResponse{
		parts:    parts,
		dirParts: dirParts,
		pdfPath:  pdfPath,
		pdfBytes: pdfBytes,
	}, nil

}
