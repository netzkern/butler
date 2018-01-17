# butler
Welcome to Butler, your personal assistent to scaffolding your projects.
Shipped with binaries for Mac, Win and Linux (64-bit).

# Usage

```sh
$ butler
```

![butler](butler.png)

# Commands

- Templating: Checkout a git project template and substituted placeholders. All template options are managed in the `butler.yml` file. Butler is shipped with a default config. If you want to create project template look [here](https://golang.org/pkg/text/template/) to get an overview about the template language.

# Development

## Installation
```
$ go get ./..
```

## Build Binaries

```
$ go get github.com/goreleaser/goreleaser
$ goreleaser
```

## Publish

```
$ git tag -a v0.1.0 -m "First release"
$ git push origin v0.1.0
```
