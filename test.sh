#!/bin/bash

set -eu

export "GOCOVERDIR=$(mktemp -d)"
trap 'rm -rf "$GOCOVERDIR"; exit 1' 1 2 3 15

go run -cover . -flag=std ./testdata/*
go tool covdata textfmt -i="$GOCOVERDIR" -o prof.out
go tool cover -html=prof.out

rm -rf "$GOCOVERDIR"
