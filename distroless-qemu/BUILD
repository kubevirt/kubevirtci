package(default_visibility = ["//visibility:private"])

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
    outs = ["qemu.tar.gz"],
    cmd = "./$(location build.sh) $(location qemu.tar.gz)",
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
    tars = [":qemu-binaries", ":passwd_tar"],
    user = "qemu",
    visibility = ["//visibility:public"],
    entrypoint = ["qemu-system-x86_64"],
)

genrule(
    name = "cirros-iso",
    srcs = [],
    outs = ["disk/cirros.img"],
    cmd = "mkdir disk && curl https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img > $@",
)

pkg_tar(
    name = "cirros-iso-tar",
    srcs = [":cirros-iso"],
    package_dir = "disk",
    owner = "107.107",
    mode = "0777",
)

container_image(
    name = "qemu_image_test",
    tars = [":qemu-binaries", ":passwd_tar", ":cirros-iso-tar"],
    user = "qemu",
    visibility = ["//visibility:public"],
    entrypoint = ["qemu-system-x86_64"],
)
