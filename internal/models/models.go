package models

import "time"

type RepoParts struct {
	Provider  string
	Owner     string
	Repo      string
	Branch    string
	Directory string
}

type DirectoryParts struct {
	Root string
	Base string
	Doc  string
}

type GithubRepoResponse struct {
	StargazersCount int       `json:"stargazers_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type RepoStats struct {
	Stars    int
	AgeYears float64
}

type PDFGenResponse struct {
	Parts    *RepoParts
	DirParts *DirectoryParts
	PdfPath  string
	PdfBytes []byte
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
	PIP PythonEnv = iota
	POETRY
	UV
)

const (
	PYTHON EnvType = iota
	NODE
)

var DocumentationName = map[DocumentationFormat]string{
	Sphinx:     "sphinx",
	MkDocs:     "mkdocs",
	Docusaurus: "docusaurus",
	GitBook:    "gitbook",
}
