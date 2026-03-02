package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bmf-san/gohan/internal/model"
)

// FileParser implements the Parser interface, reading Markdown files from disk.
// Each file may optionally begin with a YAML front matter block delimited by
// "---" lines. The remainder of the file is treated as the raw Markdown body.
//
// ExcludeFiles holds glob patterns (relative to the content directory) that
// should be skipped during ParseAll. Patterns use filepath.Match syntax.
type FileParser struct {
	excludeFiles []string
}

// NewFileParser returns a new FileParser. Pass any number of glob patterns
// (relative to the content directory) to exclude matching files from ParseAll.
func NewFileParser(excludeFiles ...string) *FileParser {
	return &FileParser{excludeFiles: excludeFiles}
}

// Parse reads the file at filePath, extracts any YAML front matter, and
// returns a fully populated *model.Article.
func (p *FileParser) Parse(filePath string) (*model.Article, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parser: read %s: %w", filePath, err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("parser: stat %s: %w", filePath, err)
	}

	fm, body, err := splitFrontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("parser: split front matter in %s: %w", filePath, err)
	}

	return &model.Article{
		FrontMatter:  fm,
		RawContent:   string(body),
		FilePath:     filePath,
		LastModified: info.ModTime(),
	}, nil
}

// ParseAll walks contentDir recursively and returns one *model.Article per
// Markdown file (.md or .markdown extension, case-insensitive).
// Files whose path (relative to contentDir) matches any pattern in
// FileParser.excludeFiles are silently skipped.
func (p *FileParser) ParseAll(contentDir string) ([]*model.Article, error) {
	var articles []*model.Article

	err := filepath.WalkDir(contentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}
		// Check exclude patterns against the path relative to contentDir.
		if len(p.excludeFiles) > 0 {
			rel, relErr := filepath.Rel(contentDir, path)
			if relErr == nil {
				for _, pattern := range p.excludeFiles {
					if matched, _ := filepath.Match(pattern, rel); matched {
						return nil
					}
				}
			}
		}
		a, parseErr := p.Parse(path)
		if parseErr != nil {
			return parseErr
		}
		articles = append(articles, a)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("parser: walk %s: %w", contentDir, err)
	}

	return articles, nil
}

// splitFrontMatter separates a YAML front matter block from the Markdown body.
// Front matter must start on the very first line as "---" and end with a
// subsequent "---" line. If no valid front matter is found the entire content
// is returned as the body unchanged.
func splitFrontMatter(data []byte) (model.FrontMatter, []byte, error) {
	var fm model.FrontMatter

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return fm, data, nil
	}

	// Find the closing "---".
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		// No closing delimiter — treat entire content as body.
		return fm, data, nil
	}

	yamlData := strings.Join(lines[1:closingIdx], "\n")
	if err := yaml.Unmarshal([]byte(yamlData), &fm); err != nil {
		return fm, nil, fmt.Errorf("unmarshal front matter: %w", err)
	}

	body := strings.Join(lines[closingIdx+1:], "\n")
	return fm, []byte(body), nil
}
