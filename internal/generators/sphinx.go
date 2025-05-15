package generators

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/jeffbrennan/pdfgen/internal/logging"
	"github.com/jeffbrennan/pdfgen/internal/models"
	"github.com/jeffbrennan/pdfgen/internal/utils"
)

func handleSphinxIssuesVersionKeyError(dirParts *models.DirectoryParts) error {
	// workaround for airflow build - should generalize after testing other sphinx builds
	extDir := "devel-common/src/sphinx_exts/"
	out, err := utils.RunCommand([]string{"ls", extDir}, dirParts.Root)
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

		filePath := dirParts.Root + "/" + extDir + file
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

func handleSphinxIssues(dirParts *models.DirectoryParts) error {
	err := handleSphinxIssuesVersionKeyError(dirParts)
	return err
}

func generateSphinxPDF(parts *models.RepoParts, dirParts *models.DirectoryParts) (string, error) {
	err := handleSphinxIssues(dirParts)
	if err != nil {
		log.Printf("Error handling Sphinx issues: %s", err)
		return "", err
	}

	// TODO: handle case where docs group does not exist
	logging.PublishLog("Generating docs as Latex...")
	out, err := utils.RunCommand([]string{
		"uv",
		"run",
		"--group",
		"docs",
		"sphinx-build",
		"-M",
		"latex",
		dirParts.Doc,
		"_build/",
	},
		dirParts.Base,
	)

	log.Printf("Sphinx build output: %s\n", out)
	if err != nil {
		log.Printf("Error running sphinx-build: %s", out)
		log.Printf("Error: %s", err)
		return "", err
	}

	outputName := parts.Repo + "_" + strings.ReplaceAll(parts.Directory, "/", "_")

	logging.PublishLog("Converting Latex to PDF...")
	out, err = utils.RunCommand(
		[]string{
			"/bin/sh",
			"-c",
			"pdflatex -interaction=nonstopmode -jobname=" + outputName + " $(find -maxdepth 1 -name '*.tex' | head -n 1)",
		},
		dirParts.Base+"/_build/latex",
	)
	log.Printf("finished running pdflatex in %s\n", dirParts.Base+"/_build/latex")

	// ignore for now
	log.Printf("uncaught pdflatex output: %s\n", out)
	log.Printf("uncaught pdflatex error: %s\n", err)

	// if err != nil {
	// 	log.Printf("Error running pdflatex: %s", out)
	// 	log.Printf("Error: %s", err)
	// 	return "", err
	// }

	pdfPath := dirParts.Base + "/_build/latex/" + outputName + ".pdf"
	log.Printf("PDF path: %s", pdfPath)
	return pdfPath, nil
}
