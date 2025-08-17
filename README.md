# Gohan - Go-based Static Site Generator

[![Go Report Card](https://goreportcard.com/badge/github.com/bmf-san/gohan)](https://goreportcard.com/report/github.com/bmf-san/gohan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Development Status](https://img.shields.io/badge/Status-In%20Development-orange)](https://github.com/users/bmf-san/projects/3)

A simple, fast, and efficient static site generator built in Go, designed for personal blogs and small team documentation sites.

## 🚀 Project Status

**Current Phase**: Foundation Development
**Target Release**: v1.0 (Q4 2025)
**Project Board**: [Gohan v1.0 Development](https://github.com/users/bmf-san/projects/3)

## 📋 Features (Planned for v1.0)

- ✅ **Simple Configuration**: Minimal YAML configuration
- ✅ **CommonMark Support**: Full CommonMark compliance with Front Matter
- ✅ **Flexible Templates**: Go html/template with custom functions
- ✅ **Fast Builds**: Optimized for speed with parallel processing
- ✅ **Code Highlighting**: Syntax highlighting for multiple languages
- ✅ **RSS/Atom Feeds**: Automatic feed generation
- ✅ **Cross-platform**: Single binary for Windows, macOS, Linux

## 🛠️ Development

### Prerequisites

- Go 1.21 or later
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/bmf-san/gohan.git
cd gohan

# Install dependencies
go mod download

# Build the project
go build -o gohan ./cmd/gohan

# Run tests
go test ./...
```

### Project Structure

```
gohan/
├── cmd/gohan/          # CLI application entry point
├── pkg/                # Public packages
├── internal/           # Private packages
├── docs/               # Documentation
├── examples/           # Example content and configs
└── themes/             # Default themes
```

## 📖 Documentation

- [Design Document](docs/design.md) - Comprehensive technical design
- [Project Management](docs/project-management.md) - Development workflow and progress tracking
- [Contributing Guide](CONTRIBUTING.md) - How to contribute to the project

## 🎯 Roadmap

### v1.0 Milestones

1. **Foundation** - Project structure and CLI framework
2. **Core Features** - Markdown parsing, templates, site generation
3. **Additional Features** - Syntax highlighting, feeds
4. **Quality & Automation** - CI/CD, documentation
5. **Release Preparation** - Testing, packaging, launch

See our [Project Board](https://github.com/users/bmf-san/projects/3) for detailed progress and current sprint information.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- Development setup
- Code style guidelines
- Pull request process
- Issue reporting

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Inspiration

Gohan is inspired by the simplicity of Go and the need for a fast, straightforward static site generator that doesn't sacrifice flexibility for simplicity.

---

**Note**: This project is currently in active development. APIs and configurations may change before the v1.0 release.