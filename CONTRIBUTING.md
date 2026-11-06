# Contributing

## Local development

Follow the requirements and quick start in README.md. Build and run with the documented platform toolchain. Source is intentionally compact and dependency-light. Keep platform-specific logic explicit and avoid introducing network services unless the feature needs one.

## Changes

1. Describe the concrete behavior being changed.
2. Keep the patch limited to that behavior.
3. Run `go test ./...` and the applicable manual checks in README.md.
4. Update usage, data-format, and limitations documentation when behavior changes.
5. Open a pull request with a short explanation and actual validation results. State skipped checks and missing platforms.

## Bug reports

Include operating system, language/runtime version, command or steps, expected result, and actual result. Supply synthetic reproduction data. Remove credentials, private screenshots, precise locations, and personal backup contents. Do not attach generated build folders or real secret environment files.
