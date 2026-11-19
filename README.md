# Chronokeep Remote

An API for remote storage of chip reads from timed races. Chronokeep's readers upload the tags
they read to it, along with heartbeats and notifications such as low battery, so reads can be
collected from anywhere.

## API

Most requests are authorized with an API key, sent as `Authorization: Bearer <key>`.
Account management uses the token from `POST /account/login` instead.

| Group | Routes |
|---|---|
| Reads | `GET /reads`, `POST /reads/add`, `DELETE /reads/delete` |
| Readers | `GET /readers` |
| Heartbeats | `POST /heartbeat`, `GET /heartbeats` |
| Notifications | `POST /notifications/save`, `GET /notifications/get` |
| Accounts and keys | `/account/...`, `/key/...` |
| Health | `/health` |

## Configuration

Settings come from environment variables.

| Variable | Required | Description |
|---|---|---|
| `DB_CONNECTOR` | | `mysql` (default) or `postgres`. |
| `DB_NAME` | Yes | Database name. |
| `DB_HOST` | Yes | Database host. |
| `DB_PORT` | Yes | Database port. |
| `DB_USER` | Yes | Database user. |
| `DB_PASSWORD` | | Database password. |
| `PORT` | | Port to listen on. Defaults to `8181`. |
| `VERSION` | | `production` for production. Anything else runs in development mode, which logs request and response bodies. |
| `AUTOTLS` | | `enabled` to serve HTTPS with a Let's Encrypt certificate. |
| `DOMAIN` | With `AUTOTLS` | The domain the certificate is for. |
| `SECRET_KEY` | Yes | Signs account tokens. At least 20 characters. |
| `REFRESH_KEY` | Yes | Signs refresh tokens. At least 20 characters. |
| `ADMIN_EMAIL`, `ADMIN_NAME`, `ADMIN_PASS` | First run | Used to create the first admin account when there are no accounts. |
| `BAN_ALLOWLIST` | | IPs or CIDRs, separated by commas, that are never banned for probing (see below). |

## Running

```sh
go build
./remote
```

The database's tables are created on first run and upgraded automatically when a new
version needs it.

## Deployment

On the server the program lives in `/chronokeep/remote/` and runs as the `chronokeep-remote`
systemd service, writing its logs to `/chronokeep/remote/logs/`. `update.sh` installs the latest
GitHub release there and restarts the service. Releases only contain the program and
`version.txt`.

Root's crontab runs `update.sh` every hour, restarts the service at 00:01 each day so each
day's log is its own file, syncs the logs to Google Drive each morning, and archives the
previous month's logs on the 1st. If you use the fail2ban jail below, also reload it just
after the daily restart so it reads the new log file.

## Security

### Probe bans

Scanners regularly ask for files this API never serves, looking for something to exploit:
`.env` files, `.php` scripts, `.git` folders, WordPress and phpMyAdmin pages and so on. The first
such request from an IP gets a 404 and bans that IP for 24 hours; while banned, it gets a 403
for everything. Bans are checked before routing and request logging, and kept in memory, so a
restart clears them.

- The IP comes from the connection itself, not `X-Forwarded-For`, so it can't be faked to get
  another address banned.
- Loopback, private and link-local addresses are never banned, and neither is anything in
  `BAN_ALLOWLIST`. Add race venues and timing networks there: one infected device on a venue's
  shared Wi-Fi could otherwise get the venue's address banned, readers included.
- Each ban is logged once, as `Banned <ip> for 24h0m0s after it asked for "<path>".`

To block these IPs at the server's firewall as well, for longer and on every port, see
[deploy/fail2ban](deploy/fail2ban/README.md).

## Testing

```sh
go test ./...
```

The tests need MySQL and Postgres databases named `remote_test` on `database.lan`, with user
and password `remote_test`, and use a local `remote_test.sqlite` file for SQLite.

## License

GNU Affero General Public License v3. See [LICENSE](LICENSE).
