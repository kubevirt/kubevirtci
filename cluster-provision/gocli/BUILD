load("@bazel_tools//tools/python:toolchain.bzl", "py_runtime_pair")

py_runtime(
    name = "python2_runtime",
    interpreter_path = "/usr/bin/python2",
    python_version = "PY2",
)

py_runtime(
    name = "python3_runtime",
    interpreter_path = "/usr/bin/python3",
    python_version = "PY3",
)

py_runtime_pair(
    name = "python_runtimes",
    py2_runtime = ":python2_runtime",
    py3_runtime = ":python3_runtime",
)

toolchain(
    name = "python_toolchain",
    toolchain = ":python_runtimes",
    toolchain_type = "@bazel_tools//tools/python:toolchain_type",
)

load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

package(default_visibility = ["//visibility:public"])

load("@bazel_gazelle//:def.bzl", "gazelle")
load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "docker_push")
load("@io_bazel_rules_docker//go:image.bzl", "go_image")
load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_bundle",
)

gazelle(
    name = "gazelle",
    prefix = "kubevirt.io/kubevirtci/cluster-provision/gocli",
)

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "kubevirt.io/kubevirtci/cluster-provision/gocli",
    deps = ["//cmd:go_default_library"],
)

go_binary(
    name = "cli",
    embed = [":go_default_library"],
    pure = "on",
)

go_image(
    name = "gocli",
    binary = ":cli",
    cmd = [""],
    visibility = ["//visibility:public"],
)

container_bundle(
    name = "container-bundle",
    images = {
        "index.docker.io/kubevirtci/gocli:latest": "//:gocli",
    },
)

docker_push(
    name = "push-all",
    bundle = ":container-bundle",
)
