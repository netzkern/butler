# Development

## Installation
```
$ go get ./...
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

## Snapshots

```
$ goreleaser --snapshot --skip-validate
```