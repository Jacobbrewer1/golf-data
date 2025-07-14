load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")

def _run_codegen(package, src, config):
    native.genrule(
        name = "generate_{}_openapi".format(package),
        srcs = [
            src,
            config,
        ],
        outs = [
            "_api.gen.go",
        ],
        cmd = "{} -package {} -o {} -generate types,client,gorilla,skip-prune -config {} -templates {} {}".format(
            "$(execpath @com_github_oapi_codegen_oapi_codegen_v2//cmd/oapi-codegen)",
            package,
            "$(location _api.gen.go)",
            "$(location {})".format(config),
            "$$(dirname $(rootpath //tools/bazel/oapi-codegen/templates:additional-properties.tmpl))",
            "$(execpath {})".format(src),
        ),
        tools = [
            "@com_github_oapi_codegen_oapi_codegen_v2//cmd/oapi-codegen",
            "//tools/bazel/oapi-codegen/templates:templates",
            "//tools/bazel/oapi-codegen/templates:additional-properties.tmpl",
            "//tools/bazel/oapi-codegen/templates/gorilla:templates",
            "//tools/bazel/oapi-codegen/templates/gorilla:gorilla-interface.tmpl",
        ],
    )

# oapi_codegen generates code in package for an OpenAPI spec src.
def oapi_codegen(package, src, **kwargs):
    files = dict()

    config = kwargs.get("config", "//tools/bazel/oapi-codegen/configs:empty_config.yaml")

    # Run oapi-codegen tool.
    _run_codegen(package, src, config)
    files["api.gen.go"] = "_api.gen.go"

    # Write generated files back to the source tree, so that IDEs, etc. work.
    write_source_files(
        name = "gen_openapi",
        files = files,
        visibility = ["//visibility:public"],
    )
