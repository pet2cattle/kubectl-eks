#!/bin/bash

set -euo pipefail

echo "Running go test..."

go test ./...

echo "Running go build..."

mkdir -p dist
go build -o dist/kubectl-eks main.go
mkdir -p $HOME/local/bin
mv dist/kubectl-eks $HOME/local/bin/kubectl-eks

echo "kubectl-eks has been built and installed into $HOME/local/bin/kubectl-eks"