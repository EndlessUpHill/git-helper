# GitHelper

GitHelper is a CLI tool that simplifies complex GitHub workflows and provides AI-powered commit message generation.

## Features

- **Repository Copying**: Copy repositories between users/organizations with full history
- **Smart Commit Messages**: Generate conventional commit messages using AI
- **Conventional Commits**: Support for manual conventional commit message creation

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/EndlessUphill/git-helper.git
cd git-helper

# Install the application
make install

# Verify installation
githelper --help
```

Make sure `~/.local/bin` is in your PATH:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Configuration

Create a configuration file at `~/.githelper.yaml`:

```yaml
# Default configuration file
github_token: "your-github-token"
default_org: "your-org"
debug: false
openai_api_key: "your-openai-api-key"
```

Or use environment variables:
```bash
export GITHELPER_GITHUB_TOKEN="your-github-token"
export GITHELPER_OPENAI_API_KEY="your-openai-api-key"
```

## Usage

### Copy Repositories

Copy a repository to your account or organization:
```bash
# Copy to user account
githelper copy https://github.com/user/repo --dest newuser/newrepo

# Copy to organization
githelper copy git@github.com:user/repo.git --dest orgname/newrepo --org

# Copy with custom settings
githelper copy https://github.com/user/repo \
  --dest org/newrepo \
  --org \
  --description "My copied repo" \
  --topics "go,cli,tools" \
  --private=false \
  --wiki=false
```

### Smart Commits

Create commits with AI-generated messages:
```bash
# Stage your changes
git add .

# Generate AI commit message
githelper commit --ai

# Generate and accept without editing
githelper commit --ai --no-edit

# Manual conventional commit
githelper commit
```

The AI commit generator will:
1. Analyze your changes
2. Generate a conventional commit message
3. Open your editor for review (unless --no-edit is used)

### Manual Commits

Create conventional commits manually:
```bash
# Interactive commit type selection
githelper commit

# Specify commit type
githelper commit -t feat

# Quick commit without editing
githelper commit -t fix --no-edit
```

## Development

### Building

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Project Structure

```
.
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and configuration
│   ├── copy.go            # Repository copying
│   └── commit.go          # Commit message generation
├── internal/              # Internal packages
│   ├── ai/               # AI integration
│   ├── github/           # GitHub API client
│   └── config/           # Configuration handling
├── Makefile              # Build and development tasks
└── main.go               # Application entry point
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`githelper commit --ai`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
