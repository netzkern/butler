# butler
The NK scaffolding tool with an interactive cli. Shipped with binaries for Mac, Win and Linux (64-bit).

# Installation

```
$ go get ./..
```

# Usage

```sh
$ butler
```

![butler](butler.png)

# Commands

- Templating: Checkout a git project template and substituted placeholders.
- Jira: Create a Project (REST Api).
- Tfs: Create a Project (REST Api).

# Build Binaries

```
$ go get github.com/goreleaser/goreleaser
$ goreleaser
```