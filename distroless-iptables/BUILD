package(default_visibility = ["//visibility:private"])

load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "docker_push")
load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_bundle", "container_image",
    container_repositories = "repositories",
)

genrule(
    name = "iptables-binaries",
    srcs = ["build.sh", "Dockerfile"],
    outs = ["iptables.tar.gz"],
    cmd = "./$(location build.sh) $(location iptables.tar.gz)",
)

container_image(
    name = "iptables_image",
    tars = [":iptables-binaries"],
    user = "0",
    visibility = ["//visibility:public"],
    entrypoint = ["iptables"],
)
