# Development

## Installation

1. Install https://github.com/golang/dep
2. Run `dep ensure`

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

## Snapshots

```
$ goreleaser --snapshot --skip-validate
```