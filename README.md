# godoc2man
Generate man pages from GoDoc comments

## Usage

You can install *godoc2man*.

```sh
go install github.com/lufia/godoc2man@latest
godoc2man [options] [pkg ...]
```

Or you can execute it directly.

```sh
go run github.com/lufia/godoc2man@latest [options] [pkg ...]
```

## Options

* *-lang*: specify the language code that is used for GoDoc document
* *-flag*: generate options section from sources with static analysis
* *-dir*: specify the output directory

## Examples

*godoc2man* generates all manuals under **cmd** directory.

```sh
godoc2man ./cmd/...
```
