# Shape Up Book Downloader

A Go-based tool to download and convert Basecamp's [Shape Up](https://basecamp.com/shapeup) book into formats suitable for e-readers and offline reading. If you're looking for a PDF version, Basecamp [already provides one](https://basecamp.com/shapeup/shape-up.pdf), as well as a [Print edition](https://basecamp-goods.com/products/shapeup).

## Features

- Downloads the complete Shape Up book content
- Converts to multiple formats:
  - Single HTML file with embedded images
  - EPUB format for e-readers
- Includes table of contents
- Embeds all images

## Installation

Choose the installation method that works best for your system:

### macOS
Using Homebrew:
```bash
brew install benjaminkitt/tap/shape-up-downloader
```

### Windows

Using Scoop:
```bash
scoop bucket add shape-up https://github.com/benjaminkitt/scoop-bucket
scoop install shape-up-downloader
```

### Download Binary
Download the latest release for your platform from our [releases page](https://github.com/benjaminkitt/shape-up-downloader/releases).

### Using Nix
If you have Nix with flakes enabled, you can install directly with:

```bash
nix profile install github:benjaminkitt/shape-up-downloader
```

### Build from Source
If you have Go installed:
```bash
go install github.com/benjaminkitt/shape-up-downloader@latest
```

## Usage

To download and convert the book to an ePUB file:

```bash
shape-up --format epub
```

or to a single HTML file:

```bash
shape-up --format html
```

# Why This Tool?

While Shape Up is freely available online and as a PDF, these formats aren't ideal for e-readers or offline reading. This tool creates versions optimized for digital reading while preserving the book's content and structure.

# Development
## Prerequisites
- Go 1.23.4 or higher
- Git

## Common Commands
```bash
# Run all tests with coverage
make test

# Run linting checks
make lint

# Build the application
make build

# Clean build artifacts
make clean

# Run all quality checks (lint + test)
make check
```

## Using Nix
You can optionally use [Nix](https://nixos.org/download.html) to manage your Go environment. This will ensure you have the correct version of Go and other dependencies installed. Once you have Nix installed, you can run:

```bash
nix develop
```

This provides a consistent development environment with:

Go compiler and tools
gopls language server
golangci-lint
delve debugger
goreleaser

## Release Process
Releases are automated via GitHub Actions. To create a new release:

1. Tag the version:
```bash
git tag -a v1.0.0 -m "Release version 1.0.0"
```
2. Push the tag:
```bash
git push origin v1.0.0
```

This will trigger the release workflow which:

Builds binaries for all supported platforms
Creates a GitHub release
Updates the Homebrew formula

# Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

# License

[MIT License](./LICENSE)

# Acknowledgments

* Inspired by [Johannes Hertenstein's shape-up-downloader](https://github.com/j6s/shape-up-downloader)
* Thanks to Basecamp for making Shape Up freely available
* Original book by [Ryan Singer](https://www.linkedin.com/in/feltpresence/)
* Built with assistance from [Cody](https://sourcegraph.com/cody)

# Disclaimer

This tool is meant for personal use only. The Shape Up book content is owned by Basecamp. Please respect their copyright and terms of use.