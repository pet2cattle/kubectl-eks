#!/bin/bash

mkdir -p dist
go build -o dist/kubectl-eks main.go
mkdir -p $HOME/local/bin
mv dist/kubectl-eks $HOME/local/bin/kubectl-eks
