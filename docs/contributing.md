# Contributing to chankit

Thank you for your interest in contributing to chankit! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- A GitHub account

### Setting Up Development Environment

1. **Fork the repository**
   ```bash
   # Visit https://github.com/yourusername/chankit and click Fork
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/yourusername/chankit.git
   cd chankit
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/original/chankit.git
   ```

4. **Install dependencies**
   ```bash
   go mod download
   ```

5. **Run tests**
   ```bash
   go test -v ./...
   ```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications

### 2. Make Changes

- Write clear, concise code
- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test -v ./...

# Run tests with race detector
go test -v -race ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

### 4. Format and Lint

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run
```

### 5. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "feat: add new debounce operator

- Implements debounce with configurable duration
- Adds comprehensive tests
- Updates documentation"
```

Commit message format:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

### 6. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Guidelines

### Go Style

Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines:

- Use `gofmt` for formatting
- Follow standard Go naming conventions
- Write idiomatic Go code
- Keep functions small and focused
- Use meaningful variable names

### Testing

- Write tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests when appropriate
- Test edge cases and error conditions

Example test:

```go
func TestMap(t *testing.T) {
    tests := []struct {
        name     string
        input    []int
        fn       func(int) any
        expected []any
    }{
        {
            name:     "square numbers",
            input:    []int{1, 2, 3},
            fn:       func(x int) any { return x * x },
            expected: []any{1, 4, 9},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            result := FromSlice(ctx, tt.input).
                Map(tt.fn).
                ToSlice()

            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### Documentation

- Add godoc comments for all exported types and functions
- Include usage examples in comments
- Update README.md if adding major features
- Update API documentation in docs/

Example documentation:

```go
// Debounce waits for the specified duration of silence before emitting the last value.
// It's useful for scenarios like search-as-you-type where you want to wait for the
// user to stop typing before performing an action.
//
// Example:
//   pipeline.Debounce(300 * time.Millisecond)
func (p *Pipeline[T]) Debounce(duration time.Duration) *Pipeline[T] {
    // implementation
}
```

## Pull Request Guidelines

### Before Submitting

- [ ] Tests pass locally
- [ ] Code is formatted (`go fmt`)
- [ ] Linters pass (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### PR Description

Provide a clear description of your changes:

```markdown
## Description
Brief description of the changes

## Motivation
Why are these changes necessary?

## Changes
- List of specific changes
- What was added/modified/removed

## Testing
How were the changes tested?

## Screenshots (if applicable)
Add screenshots for UI changes
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, your PR will be merged
4. Your contribution will be in the next release!

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Code example (minimal reproduction)

### Feature Requests

When requesting features:

- Describe the problem you're trying to solve
- Explain your proposed solution
- Provide use cases and examples
- Consider implementation complexity

## Community Guidelines

- Be respectful and inclusive
- Provide constructive feedback
- Help others learn and grow

## Questions?

- Open an issue for general questions
- Join discussions in existing issues
- Check documentation first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be recognized in:
- Release notes
- Contributors list
- Project documentation

Thank you for contributing to chankit! 🎉
