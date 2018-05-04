package(default_visibility = ["//visibility:public"])

load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_tar")
load("@io_bazel_rules_docker//contrib:push-all.bzl", "docker_push")
load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_bundle", "container_image",
    container_repositories = "repositories",
)
load("@io_bazel_rules_docker//contrib:passwd.bzl", "passwd_entry", "passwd_file")

genrule(
    name = "qemu-binaries",
    srcs = ["build.sh", "Dockerfile"],
    outs = ["qemu-img", "qemu-system-x86_64"],
    cmd = "cp Dockerfile Dockerfile.qemu && ./build.sh Dockerfile.qemu \"$(location qemu-img)\"",
)

pkg_tar(
    name = "qemu_tar",
    srcs = [":qemu-binaries"],
    package_dir = "/",
    visibility = ["//visibility:public"],
)

passwd_entry(
    name = "qemu_user",
    home = "/home/qemu",
    info = "qemu",
    shell = "/bin/bash",
    uid = 107,
    username = "qemu",
)

passwd_file(
    name = "passwd",
    entries = [
        ":qemu_user",
    ],
)

pkg_tar(
    name = "passwd_tar",
    srcs = [":passwd"],
    mode = "0644",
    package_dir = "etc",
)

container_image(
    name = "qemu_image",
    tars = [":qemu_tar", ":passwd_tar"],
    user = "qemu",
    visibility = ["//visibility:private"],
    entrypoint = ["/qemu-system-x86_64"],
)
