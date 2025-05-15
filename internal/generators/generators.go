package generators

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jeffbrennan/pdfgen/internal/env"
	"github.com/jeffbrennan/pdfgen/internal/logging"
	"github.com/jeffbrennan/pdfgen/internal/models"
	"github.com/jeffbrennan/pdfgen/internal/repo"
	"github.com/jeffbrennan/pdfgen/internal/utils"
)

func HandlePdfGeneration(url string) (models.PDFGenResponse, error) {
	parts, err := repo.ParseRepoURL(url)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error parsing URL: %s", err)
	}

	err = repo.ValidateRepo(parts)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error validating repo: %s", err)
	}

	err = repo.UpdateRepo(parts)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error updating repo: %s", err)
	}

	dirParts, err := repo.ParseRepoDir(parts)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error parsing repo directory: %s", err)
	}

	docName, err := ParseDocumentationFormat(dirParts)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error parsing documentation format: %s", err)
	}

	log.Printf("Documentation format: %s\n", models.DocumentationName[docName])
	pdfPath, err := generatePDF(parts, dirParts, docName)
	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error generating PDF: %s", err)
	}

	log.Printf("reading PDF file: %s", pdfPath)
	pdfBytes, err := os.ReadFile(pdfPath)

	if err != nil {
		return models.PDFGenResponse{}, fmt.Errorf("error reading PDF file: %s", err)
	}
	logging.PublishLog("done!")

	return models.PDFGenResponse{
		Parts:    parts,
		DirParts: dirParts,
		PdfPath:  pdfPath,
		PdfBytes: pdfBytes,
	}, nil

}

func ParseDocumentationFormat(
	dirParts *models.DirectoryParts,
) (models.DocumentationFormat, error) {
	out, err := utils.RunCommand([]string{"ls", dirParts.Doc}, dirParts.Base)
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
			logging.PublishLog(sphinxMsg)

			return models.Sphinx, nil
		}

		if strings.HasSuffix(file, "mkdocs.yml") ||
			strings.HasSuffix(file, "mkdocs.yaml") {
			log.Printf("Found MkDocs documentation: %s", file)
			return models.MkDocs, nil
		}
		if strings.HasSuffix(file, "docusaurus.config.js") {
			log.Printf("Found Docusaurus documentation: %s", file)
			return models.Docusaurus, nil
		}
		if strings.HasSuffix(file, "gitbook.yml") ||
			strings.HasSuffix(file, "gitbook.yaml") {
			log.Printf("Found GitBook documentation: %s", file)
			return models.GitBook, nil
		}
	}

	return -1, fmt.Errorf("unknown documentation format")

}
func generatePDF(parts *models.RepoParts, dirParts *models.DirectoryParts, docType models.DocumentationFormat) (string, error) {
	logging.PublishLog("Generating PDF...")
	envType, err := env.ParseEnvType(dirParts)
	if err != nil {
		return "", err
	}
	if envType == models.PYTHON {
		pythonEnv, err := env.ParsePythonEnv(dirParts)
		if err != nil {
			return "", err
		}

		env.SetupPythonEnv(dirParts, pythonEnv)
	}

	switch docType {
	case models.Sphinx:
		return generateSphinxPDF(parts, dirParts)
	}

	return "", fmt.Errorf("unknown documentation format")

}
