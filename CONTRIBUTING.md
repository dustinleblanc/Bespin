# Contributing to Bespin

Thank you for considering contributing to Bespin! This document outlines the process for contributing to the project.

## Development Setup

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/Bespin.git`
3. Install dependencies:
   ```bash
   # API dependencies
   cd api
   go mod download

   # Web client dependencies
   cd ../web
   pnpm install
   ```
4. Run the application in development mode:
   ```bash
   # From the project root
   make dev
   ```

## Code Style

- Go code should follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Vue/JavaScript code should follow the ESLint configuration in the web directory
- Use the provided `.editorconfig` file to ensure consistent formatting

## Pull Request Process

1. Create a new branch for your feature or bugfix: `git checkout -b feature/your-feature-name`
2. Make your changes
3. Run tests to ensure everything works: `make test`
4. Commit your changes with a descriptive commit message
5. Push your branch to your fork: `git push origin feature/your-feature-name`
6. Open a pull request against the main repository

## Commit Message Guidelines

Please follow these guidelines for commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

## License

By contributing to Bespin, you agree that your contributions will be licensed under the project's MIT License.
