# butler
The NK scaffolding tool with an interactive prompts. Shipped with binaries for Mac, Win and Linux (64-bit).

# Installation

```
$ go get ./..
```

# Usage

```sh
$ butler
```

# Commands

- Templating: Checkout a git project template and substituted placeholders.
- Jira: Create a Project (REST Api).
- Tfs: Create a Project (REST Api).

# Build Binaries

```
$ go get github.com/goreleaser/goreleaser
$ goreleaser
```