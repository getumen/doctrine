#!/bin/sh

set -eu

protoc -I=proto --go_out=phalanxpb --go_opt=paths=source_relative proto/command.proto
