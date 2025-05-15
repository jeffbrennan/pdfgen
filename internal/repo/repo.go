package repo

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/jeffbrennan/pdfgen/internal/logging"
	"github.com/jeffbrennan/pdfgen/internal/models"
	"github.com/jeffbrennan/pdfgen/internal/utils"
)

func GetGithubAPIResponseRepo(owner string, repo string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%v/%v", owner, repo)
	log.Printf("requesting a response from %s...", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	token, err := utils.LoadSecret("GITHUB_TOKEN")
	if err != nil {
		log.Fatal("GITHUB_TOKEN not set")
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
func UpdateRepo(parts *models.RepoParts) error {
	// pulls or clones depending on if the repo exists
	baseDir := "./repos"
	targetDir := fmt.Sprintf(
		"%s/%s", baseDir, parts.Repo,
	)

	updateRepoMsg := fmt.Sprintf(
		"Updating %s/%s/%s...",
		parts.Provider,
		parts.Owner,
		parts.Repo,
	)

	log.Print(updateRepoMsg)
	logging.PublishLog(updateRepoMsg)

	_, err := utils.RunCommand([]string{"ls"}, targetDir)
	if err == nil {
		log.Printf("Directory %s already exists", targetDir)
		err = pullRepo(targetDir)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = utils.RunCommand([]string{"mkdir", "-p", baseDir}, "")
	if err != nil {
		return err
	}

	return cloneRepo(parts, baseDir)
}

func pullRepo(targetDir string) error {
	_, err := utils.RunCommand([]string{"git", "pull"}, targetDir)
	return err
}

func cloneRepo(parts *models.RepoParts, baseDir string) error {
	baseURL := fmt.Sprintf(
		"https://%s/%s/%s.git",
		parts.Provider,
		parts.Owner,
		parts.Repo,
	)
	_, err := utils.RunCommand(
		[]string{
			"git",
			"clone",
			baseURL,
			"-b",
			parts.Branch,
			"--single-branch",
			"--depth",
			"1",
		},
		baseDir,
	)
	return err
}
