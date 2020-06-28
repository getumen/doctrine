load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = glob(["*.go"], exclude=["*_test.go"]) + glob(["*.h"]) + glob(["*.c"]),
    importpath = "github.com/tecbot/gorocksdb",
    cgo = True,
    visibility = ["//visibility:public"],
    cdeps = [
        "@com_github_facebook_rocksdb//:librocksdb",
    ],
    clinkopts = [
        "-lstdc++",
        "-lm",
        "-ldl",
    ]
)