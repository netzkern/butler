<p align="center">
<img src="https://raw.githubusercontent.com/netzkern/butler/master/logo.png" alt="butler" style="max-width:100%;">
</p>

Welcome to Butler, your personal assistent to scaffolding your projects.
Shipped with binaries for Mac, Win and Linux (64-bit).

# Usage

[Download here](https://github.com/netzkern/butler/releases)

```sh
$ butler
```

# Commands

## Project Templates

This command will create a new project template. All available project templates are managed in the `butler.yml` file. Butler is shipped with a default config. If you want to create a project template look [here](https://golang.org/pkg/text/template/) for the template features. We use a unique delimiter to avoid collsion with existing template engines.

### Delimiter

```
butler{ .Project.Name }
```

### Available variables:

- `Project.Name`: Project name
- `Project.Description`: Project description
- `Date`: Current Date (RFC3339)
- `Year`: Current year

You can specify custom variables in the `butler.yml` file. They can be accessed e.g `butler{ .Vars.company }`

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
