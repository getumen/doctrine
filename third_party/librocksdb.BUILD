load("@rules_cc//cc:defs.bzl", "cc_library")
licenses(["notice"])

genrule(
    name = "build_version",
    srcs = glob([".git/**/*"]) + [
        "util/build_version.cc.in",
    ],
    outs = [
        "util/build_version.cc",
    ],
    # `git rev-parse HEAD` returns the SHA, but it does not work in all the environments
    # when combined with bazel sandboxing
    cmd = "GIT_DIR=external/com_github_facebook_rocksdb/.git; " +
        "GIT_SHA=$$(cat $$GIT_DIR/HEAD | cut -d \" \" -f 2 | xargs -I {} cat $$GIT_DIR/{} || echo 'v5.7.3'); " +
        "sed -e s/@@GIT_SHA@@/$$GIT_SHA/ -e s/@@GIT_DATE_TIME@@/$$(date +%F)/ " +
        "external/com_github_facebook_rocksdb/util/build_version.cc.in >> $(@)",
)

cc_library(
    name = "librocksdb",
    srcs = glob([
        "**/*.h",
    ], exclude=[
        "db_stress_tool/**/*.h",
        "port/win/*.h",
        "third-party/gtest-1.7.0/fused-src/gtest/*.h",
        "include/rocksdb/utilities/env_librados.h",
    ]) + [
        ":build_version",
    ] + glob([
        "**/*.cc"
    ], exclude=[
        "**/*_test.cc",
        "**/*_bench.cc",
        "java/**/*.cc",
        "db_stress_tool/**/*.cc",
        "examples/**/*.cc",
        "port/win/*.cc",
        "tools/**/*.cc",
        "third-party/gtest-1.7.0/fused-src/gtest/*.cc",
        "utilities/env_librados.cc",
        "db/db_test2.cc",
    ]),
    hdrs = glob([
        "include/rocksdb/**/*.h",
    ], exclude=[
        "include/rocksdb/utilities/env_librados.h",
    ]),
    includes = [
        ".",
        "include",
        "util",
    ],
    defines = [
        "ROCKSDB_FALLOCATE_PRESENT",
        "ROCKSDB_JEMALLOC",
        "ROCKSDB_LIB_IO_POSIX",
        "ROCKSDB_MALLOC_USABLE_SIZE",
        "ROCKSDB_PLATFORM_POSIX",
        "ROCKSDB_SUPPORT_THREAD_LOCAL",
    ],
    copts = [
        "-DGFLAGS=gflags",
        "-DJEMALLOC_NO_DEMANGLE",
        "-DOS_LINUX",
        "-DSNAPPY",
        "-DHAVE_SSE42",
        "-DZLIB",
        "-fno-omit-frame-pointer",
        "-momit-leaf-frame-pointer",
        "-msse4.2",
        "-pthread",
        "-Werror",
        "-Wno-sign-compare",
        "-Wshadow",
        "-Wno-unused-parameter",
        "-Wno-unused-variable",
        "-Woverloaded-virtual",
        "-Wnon-virtual-dtor",
        "-Wno-missing-field-initializers",
        "-std=c++11",
    ],
    linkopts = [
        "-lm",
        "-lpthread",
    ],
    deps = [
        "@//third_party/libbz2:libbz2",
        "@com_github_gflags_gflags//:gflags",
        "@//third_party/libsnappy:libsnappy",
        "@//third_party/libz:libz",
        "@//third_party/liblz4:liblz4",
        "@//third_party/libzstd:libzstd",
        "@com_google_googletest//:gtest",
        "@//third_party/libjemalloc:libjemalloc",
    ],
    visibility = ["//visibility:public"],
)