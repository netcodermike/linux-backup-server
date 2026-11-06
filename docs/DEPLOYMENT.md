# Linux systemd deployment

1. Build `go build -o vault .` on your Linux host, or cross-compile with the correct GOOS/GOARCH.
2. Create a dedicated `vault` system user/group without login access and `/var/lib/vault` owned by that account with mode 0700.
3. Install the executable as `/usr/local/bin/vault`, owned by root and executable by the service user.
4. Copy `deploy/vault.env.example` to `/etc/vault.env`. Generate and replace BACKUP_TOKEN using `openssl rand -hex 32`; set file mode 0600. Keep BACKUP_KEY only on clients, in a secure password manager or secret store.
5. Install `deploy/vault.service` into `/etc/systemd/system/`, run `systemctl daemon-reload`, then `systemctl enable --now vault` as the administrator of your server.
6. Inspect `systemctl status vault` and `journalctl -u vault`. Back up a synthetic folder and restore it before relying on the service.

The listener defaults to 127.0.0.1:8088. For remote use, configure a maintained TLS reverse proxy, limit request sizes to the documented object size, and retain the Authorization header. Avoid putting tokens in URLs or logs. Browser crypto requires HTTPS or localhost. Restrict network exposure with your server firewall. Provide disk quotas/monitoring externally. The service never needs outbound connections.

Back up the entire data directory to a separate disk. Protect the client encryption key independently. Restore-test regularly. A single shared vault uses one encryption key; create a separate vault/data directory/token for each unrelated key. Stop the service before copying storage for a consistent backup. Deleting individual objects can invalidate many snapshots because of deduplication.
