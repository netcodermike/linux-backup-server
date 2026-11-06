# Validation record

Authoring host: Windows, November 6, 2026. Requirements: Go 1.22+; Linux server with systemd; CLI also supports Windows/macOS.

The README verification command is `go test ./...`. The companion GitHub Actions workflow describes reproducible build/test checks. Merely including that workflow does not mean it ran successfully. Host-specific verification results are recorded before publication.

Manual acceptance steps are in README.md. GUI behavior, hardware sensors, permissions, audio output, and OS isolation need verification on their target platform; compilation alone cannot establish these properties.

## Authoring result

Target toolchain/platform unavailable on the Windows authoring host. Source and build configuration were reviewed, and a target-platform GitHub Actions build/check workflow is provided. No successful native build or runtime result is claimed yet.
