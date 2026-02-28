// Package parser parses Markdown content and YAML Front Matter from content files.
package parser

import (
	"github.com/bmf-san/gohan/internal/model"
)

// Parser converts a raw Markdown file (with optional YAML front matter) into
// a model.Article ready for further processing.
type Parser interface {
	// Parse reads the file at filePath and returns a fully populated Article.
	Parse(filePath string) (*model.Article, error)

	// ParseAll walks contentDir and returns one Article per Markdown file found.
	ParseAll(contentDir string) ([]*model.Article, error)
}
