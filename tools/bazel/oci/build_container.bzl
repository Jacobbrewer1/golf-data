load("@rules_oci//oci:defs.bzl", "oci_image", "oci_tarball")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")

# build_container builds an OCI-compliant container image, based off of the go_binary target supplied in the name argument.
def build_container(name):
    build_container_with_extras(name, [])

# build_container builds an OCI-compliant container image, based off of the go_binary target supplied in the name argument.
# It also includes any extra tarballs supplied in the extra_tars argument.
def build_container_with_extras(name, extra_tars):
    pkg_tar(
        name = "tar",
        srcs = [":{}".format(name)],
        tags = ["oci"],
    )

    oci_image(
        name = "image",
        base = select({
            "//conditions:default": "@ubuntu_base",
        }),
        entrypoint = ["/{}".format(name)],
        tars = extra_tars + [":tar"],
        tags = ["oci"],
    )

    oci_tarball(
        name = "oci.tar",
        image = ":image",
        repo_tags = ["{}:latest".format(name)],
        tags = ["oci"],
    )
