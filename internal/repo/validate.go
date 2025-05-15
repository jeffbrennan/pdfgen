package repo

import (
	"fmt"
	"log"

	"github.com/jeffbrennan/pdfgen/internal/models"
)

func ValidateRepo(parts *models.RepoParts) error {
	// guard against improper usage only accepting large, established projects
	minNumStars := 100
	minNumStarsNewRepo := 1000
	minRepoAgeYears := 1.0

	response, err := GetGithubAPIResponseRepo(parts.Owner, parts.Repo)
	if err != nil {
		return err
	}

	repoStats, err := ParseGithubAPIResponse(response)
	if err != nil {
		return err
	}

	if repoStats.Stars < minNumStars {
		return fmt.Errorf("repo has less than %d stars", minNumStars)
	}

	if (repoStats.AgeYears < minRepoAgeYears) && (repoStats.Stars < minNumStarsNewRepo) {
		return fmt.Errorf("repo is less than %f years old and has less than %d stars", minRepoAgeYears, minNumStarsNewRepo)
	}

	log.Printf("%s/%s is valid", parts.Owner, parts.Repo)
	return nil
}
