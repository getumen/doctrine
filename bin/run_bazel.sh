#!/bin/sh

set -eu

bazel run //:gazelle
bazel run //:gazelle -- update-repos -from_file=phalanx/go.mod -to_macro=repositories.bzl%go_repositories -prune=true
bazel build //...
bazel test //...
