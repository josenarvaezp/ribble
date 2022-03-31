#!/bin/sh

go build -ldflags "-s -w" -o "$1"/gen_job  "$2"
chmod +x "$1"/gen_job