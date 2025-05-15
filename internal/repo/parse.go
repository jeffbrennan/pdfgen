package repo

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jeffbrennan/pdfgen/internal/models"
)

func ParseGithubAPIResponse(responseBody []byte) (*models.RepoStats, error) {
	var response models.GithubRepoResponse
	err := json.Unmarshal(responseBody, &response)
	if err != nil {
		return &models.RepoStats{}, err
	}
	ageYears := time.Since(response.CreatedAt).Hours() / (24 * 365.25)

	return &models.RepoStats{
		Stars:    response.StargazersCount,
		AgeYears: ageYears,
	}, nil

}

func ParseRepoDir(parts *models.RepoParts) (*models.DirectoryParts, error) {
	var docDir string
	var baseDir string

	rootDir := fmt.Sprintf(
		"%s/%s", "./repos", parts.Repo,
	)

	if parts.Directory == "" {
		baseDir = rootDir
		docDir = "docs/"
	} else {

		docParts := strings.Split(parts.Directory, "/")
		// strip last element
		baseDir = rootDir + "/" + strings.Join(docParts[:len(docParts)-1], "/")
		log.Printf("Base dir: %s", baseDir)

		docDir = docParts[len(docParts)-1] + "/"
	}

	dirParts := &models.DirectoryParts{
		Root: rootDir, //  .../airflow
		Base: baseDir, // .../airflow/airflow-core/
		Doc:  docDir,  // docs/
	}

	return dirParts, nil
}
func ParseRepoURL(url string) (*models.RepoParts, error) {
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
		return &models.RepoParts{
			Provider: provider,
			Owner:    owner,
			Repo:     repo,
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
		return &models.RepoParts{
			Provider:  provider,
			Owner:     owner,
			Repo:      repo,
			Branch:    branch,
			Directory: "docs",
		}, nil

	}

	directory := strings.Join(parts[4+branchPadding:], "/")

	return &models.RepoParts{
		Provider:  provider,
		Owner:     owner,
		Repo:      repo,
		Branch:    branch,
		Directory: directory,
	}, nil

}
