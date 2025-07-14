load("@aspect_bazel_lib//lib:write_source_files.bzl", "write_source_files")
load("@io_bazel_rules_go//go:def.bzl", "gomock")

def mock(library, interfaces = []):
    gomock(
        name = "generate_mocks",
        out = "_mock.gen.go",
        interfaces = interfaces,
        library = library,
        package = "mock",
        mockgen_tool = "@org_uber_go_mock//mockgen",
        mockgen_model_library = "@org_uber_go_mock//mockgen/model",
    )

    # strip_command strips the header off of the top of the file, as the header has dynamic content, which
    # causes tests to fail in CI. We should provide '-write_command_comment=false' to mockgen, but this flag
    # isn't available in rules_go#gomock.
    native.genrule(
        name = "strip_command",
        srcs = [
            "_mock.gen.go",
        ],
        outs = [
            "_mock_stripped.gen.go",
        ],
        cmd = "tail -n +9 $(location _mock.gen.go) | sed -e 's/GoMock package.$$/GoMock package. DO NOT EDIT/g' > $(location _mock_stripped.gen.go)",
    )

    # Write generated files back to the source tree, so that IDEs, etc. work.
    write_source_files(
        name = "gen_mock",
        files = {
            "mock.gen.go": "_mock_stripped.gen.go",
        },
        visibility = ["//visibility:public"],
    )
