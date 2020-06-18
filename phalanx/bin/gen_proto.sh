#!/bin/sh

set -eu

protoc -I=phalanxpb --go_out=phalanxpb --go_opt=paths=source_relative phalanxpb/command.proto
