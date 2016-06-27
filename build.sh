#!/usr/bin/env bash

GOOS=linux GOARCH=amd64 go build -o ./zdd_linux_adm64
tar -czf ./zdd_linux_amd64.tar.gz ./zdd_linux_adm64
