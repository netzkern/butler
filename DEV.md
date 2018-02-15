# Development

## Installation

```shell
go get ./...
```

## Build Binaries

```shell
go get github.com/goreleaser/goreleaser
goreleaser
```

## Publish

```shell
git tag -a v0.1.0 -m "First release"
git push origin v0.1.0
```

## Snapshots

```shell
goreleaser --snapshot --skip-validate
```
