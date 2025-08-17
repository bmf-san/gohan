# Contributing to Gohan

Thank you for your interest in contributing to Gohan! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Issue Guidelines](#issue-guidelines)
- [Pull Request Guidelines](#pull-request-guidelines)

## 🤝 Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## 🚀 Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/gohan.git
   cd gohan
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/bmf-san/gohan.git
   ```

## 🛠 Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build the project**:
   ```bash
   go build -o gohan ./cmd/gohan
   ```

3. **Run tests**:
   ```bash
   go test ./...
   ```

4. **Run with test data**:
   ```bash
   ./gohan build --config=examples/basic/config.yaml
   ```

## 📝 Making Changes

### Before You Start

1. **Check existing issues** to see if your change is already being worked on
2. **Create an issue** for significant changes to discuss the approach
3. **Keep changes focused** - one feature/fix per PR

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our code style guidelines

3. **Add tests** for new functionality

4. **Update documentation** if needed

5. **Test your changes**:
   ```bash
   go test ./...
   go build ./...
   ```

6. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

## 📤 Submitting Changes

### Preparing Your Pull Request

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** from your fork to the main repository

### Pull Request Requirements

- [ ] Clear title and description
- [ ] Related issue linked
- [ ] All tests passing
- [ ] Code follows style guidelines
- [ ] Documentation updated if needed
- [ ] No merge conflicts

## 🎨 Code Style

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use `golint` and `go vet` for linting
- Write clear, descriptive variable and function names
- Add comments for exported functions and complex logic

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

**Examples:**
- `feat: add support for custom templates`
- `fix: resolve issue with markdown parsing`
- `docs: update installation instructions`

## 🧪 Testing

### Writing Tests

- Write unit tests for all new functionality
- Use Go's built-in testing framework
- Follow the naming convention: `TestFunctionName`
- Place tests in the same package as the code being tested

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific test
go test -run TestSpecificFunction ./pkg/specific
```

### Test Coverage

- Aim for at least 80% test coverage for new code
- Check coverage with: `go test -cover ./...`

## 📚 Documentation

### What to Document

- All exported functions and types
- Configuration options
- CLI commands and flags
- Examples and tutorials
- Installation instructions

### Documentation Format

- Use clear, concise language
- Provide examples where helpful
- Keep documentation up to date with code changes
- Use proper Markdown formatting

## 🐛 Issue Guidelines

### Before Creating an Issue

1. **Search existing issues** to avoid duplicates
2. **Check documentation** for answers
3. **Use the latest version** of Gohan

### Creating Good Issues

- **Use appropriate templates** (bug report, feature request, etc.)
- **Provide detailed information** including steps to reproduce
- **Include relevant configuration** and environment details
- **Use clear, descriptive titles**

## 🔄 Pull Request Guidelines

### Before Submitting

- [ ] Issue exists and is linked
- [ ] Branch is up to date with main
- [ ] All tests pass
- [ ] Code is properly formatted
- [ ] Documentation is updated
- [ ] Commit messages follow conventions

### PR Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainers
3. **Testing** on different environments if needed
4. **Approval** and merge by maintainers

### After Merge

- **Delete your feature branch**
- **Update your local main branch**
- **Consider helping with related issues**

## 🏷 Release Process

Releases are managed by maintainers using semantic versioning:

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backwards compatible
- **Patch** (0.0.X): Bug fixes, backwards compatible

## 🆘 Getting Help

- **Discord**: [Join our community](https://discord.gg/gohan) (if applicable)
- **Discussions**: Use GitHub Discussions for questions
- **Issues**: Create an issue for bugs or feature requests
- **Email**: Contact maintainers directly for sensitive issues

## 🙏 Recognition

Contributors are recognized in:

- Release notes
- README contributors section
- Special mentions for significant contributions

Thank you for contributing to Gohan! 🎉
