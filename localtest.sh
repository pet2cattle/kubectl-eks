#!/bin/bash

go build -o dist/kubectl-eks main.go
mv dist/kubectl-eks $HOME/local/bin/kubectl-eks