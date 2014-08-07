#!/bin/sh

ragel -G2 -Z -o scan_auto.go ragel/exec.rl
#go tool yacc -p yy -o parse_auto.go yacc/parser.y

go fmt
go test
