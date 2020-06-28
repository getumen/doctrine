workspace(name = "doctrine")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository", "new_git_repository")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "0c10738a488239750dbf35336be13252bad7c74348f867d30c3c3e0001906096",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.2/rules_go-v0.23.2.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.2/rules_go-v0.23.2.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

# Group the sources of the library so that CMake rule have access to it
all_content = """filegroup(name = "all", srcs = glob(["**"]), visibility = ["//visibility:public"])"""

# Rule repository
http_archive(
    name = "rules_foreign_cc",
    strip_prefix = "rules_foreign_cc-master",
    url = "https://github.com/bazelbuild/rules_foreign_cc/archive/master.zip",
    sha256 = "3b21a34d803f2355632434865c39d122a57bf3bf8bb2636e27b474aeac455e5c",
)

load("@rules_foreign_cc//:workspace_definitions.bzl", "rules_foreign_cc_dependencies")

rules_foreign_cc_dependencies()

http_archive(
    name = "com_github_gflags_gflags",  # match the name defined in its WORKSPACE file
    url = "https://github.com/gflags/gflags/archive/v2.2.2.tar.gz",
    strip_prefix = "gflags-2.2.2",
    sha256 = "34af2f15cf7367513b352bdcd2493ab14ce43692d2dcd9dfc499492966c64dcf",
)

http_archive(
    name = "com_github_madler_zlib",
    url = "https://github.com/madler/zlib/archive/v1.2.11.tar.gz",
    strip_prefix = "zlib-1.2.11",
    build_file_content = all_content,
    sha256 = "629380c90a77b964d896ed37163f5c3a34f6e6d897311f1df2a7016355c45eff",
)

http_archive(
    name = "com_github_enthought_bzip2",
    url = "https://github.com/enthought/bzip2/archive/v1.0.6.tar.gz",
    strip_prefix = "bzip2-1.0.6",
    build_file_content = all_content,
    sha256 = "595d20b4081e08ae63ec42ed8c2829ac86846745628d05a84c484c43cb1df09d",
)

http_archive(
    name = "com_github_lz4_lz4",
    url = "https://github.com/lz4/lz4/archive/v1.9.2.tar.gz",
    strip_prefix = "lz4-1.9.2",
    build_file_content = all_content,
    sha256 = "658ba6191fa44c92280d4aa2c271b0f4fbc0e34d249578dd05e50e76d0e5efcc",
)

http_archive(
    name = "com_github_google_snappy",
    url = "https://github.com/google/snappy/archive/1.1.8.tar.gz",
    strip_prefix = "snappy-1.1.8",
    build_file_content = all_content,
    sha256 = "16b677f07832a612b0836178db7f374e414f94657c138e6993cbfc5dcc58651f",
)

http_archive(
    name = "com_github_facebook_zstd",
    url = "https://github.com/facebook/zstd/archive/v1.4.5.tar.gz",
    strip_prefix = "zstd-1.4.5",
    build_file_content = all_content,
    sha256 = "734d1f565c42f691f8420c8d06783ad818060fc390dee43ae0a89f86d0a4f8c2",
)

git_repository(
    name = "com_google_googletest",
    remote = "https://github.com/google/googletest",
    commit = "dea0216d0c6bc5e63cf5f6c8651cd268668032ec",
    shallow_since = "1548823078 -0500",
)

http_archive(
    name = "com_github_jemalloc_jemalloc",
    url = "https://github.com/jemalloc/jemalloc/archive/5.2.1.tar.gz",
    strip_prefix = "jemalloc-5.2.1",
    build_file_content = all_content,
    sha256 = "ed51b0b37098af4ca6ed31c22324635263f8ad6471889e0592a9c0dba9136aea",
)

http_archive(
    name = "com_github_facebook_rocksdb",
    url = "https://github.com/facebook/rocksdb/archive/v5.18.3.tar.gz",
    strip_prefix = "rocksdb-5.18.3",
    build_file = "@//:third_party/librocksdb.BUILD",
    sha256 = "7fb6738263d3f2b360d7468cf2ebe333f3109f3ba1ff80115abd145d75287254",
)

http_archive(
    name = "com_github_tecbot_gorocksdb",
    url = "https://github.com/tecbot/gorocksdb/archive/v5.0.tar.gz",
    strip_prefix = "gorocksdb-5.0",
    build_file = "@//:third_party/gorocksdb.BUILD",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

load("//:repositories.bzl", "go_repositories")

# gazelle:repository_macro repositories.bzl%go_repositories
go_repositories()

http_archive(
    name = "rules_python",
    url = "https://github.com/bazelbuild/rules_python/releases/download/0.0.2/rules_python-0.0.2.tar.gz",
    strip_prefix = "rules_python-0.0.2",
    sha256 = "b5668cde8bb6e3515057ef465a35ad712214962f0b3a314e551204266c7be90c",
)

load("@rules_python//python:repositories.bzl", "py_repositories")

py_repositories()

# Only needed if using the packaging rules.
load("@rules_python//python:pip.bzl", "pip_repositories")

pip_repositories()

git_repository(
    name = "com_google_protobuf",
    commit = "09745575a923640154bcf307fba8aedff47f240a",
    remote = "https://github.com/protocolbuffers/protobuf",
    shallow_since = "1558721209 -0700",
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()
