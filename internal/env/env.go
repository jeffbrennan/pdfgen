package env

import (
	"fmt"
	"log"
	"strings"

	"github.com/jeffbrennan/pdfgen/internal/logging"
	"github.com/jeffbrennan/pdfgen/internal/models"
	"github.com/jeffbrennan/pdfgen/internal/utils"
)

func setupPythonEnvPip(dirParts *models.DirectoryParts) error {

	_, err := utils.RunCommand(
		[]string{"uv", "pip", "install", "-r", "requirements.txt"},
		dirParts.Base,
	)
	return err

}

func setupPythonEnvPoetry(dirParts *models.DirectoryParts) error {
	// TODO: parse pyproject.toml to look for a docs group
	_, err := utils.RunCommand(
		[]string{"uvx", "migrate-to-uv"},
		dirParts.Base,
	)

	if err != nil {
		return err
	}

	_, err = utils.RunCommand(
		[]string{"uv", "sync"},
		dirParts.Base,
	)

	return err
}

func setupPythonEnvUV(dirParts *models.DirectoryParts) error {
	_, err := utils.RunCommand(
		[]string{"uv", "sync"},
		dirParts.Base,
	)

	return err
}

func SetupPythonEnv(dirParts *models.DirectoryParts, env models.PythonEnv) error {
	logging.PublishLog("Setting up Python environment...")
	_, err := utils.RunCommand(
		[]string{"uv", "venv"},
		dirParts.Base,
	)

	if err != nil {
		return err
	}

	switch env {
	case models.PIP:
		return setupPythonEnvPip(dirParts)
	case models.POETRY:
		return setupPythonEnvPoetry(dirParts)
	case models.UV:
		return setupPythonEnvUV(dirParts)
	default:
		return fmt.Errorf("unknown python env")
	}
}

func setupNodeEnv(dirParts *models.DirectoryParts) error {
	return fmt.Errorf("node env setup not implemented")
}

func ParseEnvType(dirParts *models.DirectoryParts) (models.EnvType, error) {
	out, err := utils.RunCommand([]string{"ls"}, dirParts.Base)
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "package.json") {
			log.Printf("Found node env: %s", file)
			return models.NODE, nil
		}
		if strings.HasSuffix(file, "requirements.txt") ||
			strings.HasSuffix(file, "poetry.lock") ||
			strings.HasSuffix(file, "uv.lock") ||
			strings.HasSuffix(file, "pyproject.toml") {
			log.Printf("Found python env: %s", file)
			return models.PYTHON, nil
		}
	}

	return -1, fmt.Errorf("unknown env")
}

func ParsePythonEnv(dirParts *models.DirectoryParts) (models.PythonEnv, error) {
	logging.PublishLog("Parsing Python env...")
	out, err := utils.RunCommand([]string{"ls"}, dirParts.Base)
	if err != nil {
		return -1, err
	}

	files := strings.Split(string(out), "\n")
	for _, file := range files {
		if strings.HasSuffix(file, "requirements.txt") {
			log.Printf("Found pip env: %s", file)
			return models.PIP, nil
		}

		if strings.HasSuffix(file, "uv.lock") {
			log.Printf("Found uv: %s", file)
			return models.UV, nil
		}
		if strings.HasSuffix(file, "poetry.lock") || strings.HasSuffix(file, "pyproject.toml") {
			return models.POETRY, nil
		}
	}

	return -1, fmt.Errorf("unknown env")

}
