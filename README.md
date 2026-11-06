# Linux Backup Vault

[![Build and verify](https://github.com/netcodermike/linux-backup-server/actions/workflows/verify.yml/badge.svg)](https://github.com/netcodermike/linux-backup-server/actions/workflows/verify.yml)


Store client-encrypted folder backups in a self-hosted content-addressed vault, list snapshots, and restore files with verified hashes.

**Language:** Go 1.22+  
**Platform:** Linux server with systemd; CLI also supports Windows/macOS  
**Status:** First-release learning project. See implemented scope and limitations below; this is not a production-completeness claim.

## Features

- AES-256-GCM encrypts file objects and manifests on the client.
- SHA-256 object IDs deduplicate identical file contents across snapshots sharing a key.
- Bearer authentication and constant-time token comparison.
- Immutable object IDs, write-by-rename persistence, and restore hash verification.
- A hardened sample systemd service, environment example, and browser restore interface.

## Requirements and quick start

Clone the repository, enter its folder, and install the language/platform tools listed above. Commands run from the repository root. No credentials are required except for Backup Vault's explicitly generated secrets.

```sh
git clone https://github.com/netcodermike/linux-backup-server.git
cd linux-backup-server
go build -o vault .
# Linux/macOS shell: keep these generated values in your password manager.
export BACKUP_TOKEN="$(openssl rand -hex 32)"
export BACKUP_KEY="$(openssl rand -hex 32)"
./vault serve
# In a second shell export the SAME token/key, then:
./vault backup http://127.0.0.1:8088 /path/to/folder
./vault list http://127.0.0.1:8088
./vault restore http://127.0.0.1:8088 SNAPSHOT_ID /path/to/new-folder
```

## Verification

```sh
go test ./...
```

Start a temporary vault, back up a folder, change a file, and back up again. Restore both IDs to separate fresh folders and compare bytes. Confirm unchanged files occupy one server object. Try a wrong token and wrong key. In the browser UI, load snapshot IDs with your token, decrypt a manifest with your key, and download an original file.

See [validation notes](docs/VALIDATION.md) for what was actually executed during authoring. A committed CI workflow is a future check until its run is observed successful.

## Architecture and data

`main.go` contains server routes, HTTP client, AES-GCM helpers, and backup/restore commands. `main_test.go` tests encryption, key validation, and tamper rejection. The server stores `objects/HASH` and `manifests/ID` beneath `BACKUP_DIR` (default `data`). It never needs the encryption key. Manifests encrypt relative-path-to-hash maps. Every encryption operation uses a fresh random nonce, prefixed to ciphertext. The existing ciphertext for a content ID is reused by the server.

`BACKUP_TOKEN` must be at least 24 characters. Clients need `BACKUP_KEY`, a random 32-byte key encoded as 64 hex characters. `BACKUP_ADDR` defaults to `127.0.0.1:8088`. The API accepts authenticated GET/PUT for `/objects/HASH` and `/manifests/ID`, plus GET `/manifests`. Files are limited to 64 MiB each; HTTP objects to 65 MiB. Restores require a new destination folder whose parent exists. Paths and decrypted object hashes are validated.

## Limitations and next steps

Do not lose the key: it cannot be recovered from the server. Use one encryption key per vault; uploading identical content with another key would reuse incompatible ciphertext. Plaintext hashes and ciphertext sizes leak equality and approximate size. This is not a zero-knowledge or multi-tenant design. No quotas, garbage collection, chunking, incremental metadata scan, key rotation, scheduling, or directory metadata preservation. Backups are not transactionally consistent with actively changing files. Only regular files are copied; symlinks and special files are skipped. Production deployment requires TLS at a reverse proxy, protected token storage, disk monitoring, and tested restores. The sample server defaults to local access.

## Troubleshooting

401 means token mismatch. AES authentication failure usually means the wrong key or damaged ciphertext. Existing restore directories are rejected. A too-large file aborts the backup before publishing its manifest; uploaded orphan objects can remain. `data` must be writable by the service user. For systemd setup read `docs/DEPLOYMENT.md`. Never commit real environment files or secrets.

## Development

Read the source entry point described above, make a focused change, run the verification command, and exercise the relevant manual scenario. [Contributing](CONTRIBUTING.md) explains bug reports and pull requests; [security notes](SECURITY.md) describe the trust assumptions. Runtime/generated files are ignored. No telemetry or paid API dependencies are included.

## License

[MIT](LICENSE), copyright 2026 netcodermike.
