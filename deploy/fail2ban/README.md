# Blocking probing IPs with fail2ban

Chronokeep Remote bans IPs that ask for paths only an exploit scanner would ask for (`.env`,
`.php`, `.git`, `wp-admin` and so on). That ban lives in the API's memory for 24 hours and is
cleared when the service restarts. fail2ban takes the same IPs and blocks them in the
server's firewall instead, so:

- their traffic is dropped before it reaches the API,
- bans last longer (7 days, growing for repeat offenders) and survive restarts,
- they're blocked on every port, not just the API's.

DigitalOcean's Cloud Firewalls can't do this, since they only have allow rules. fail2ban runs
on the droplet itself.

## Install

Releases only contain the program, so fetch these files from the repository. On the server,
as root:

```sh
apt install fail2ban nftables
BASE=https://raw.githubusercontent.com/grecaun/chronokeep-remote/main/deploy/fail2ban
curl -fL $BASE/filter.d/chronokeep-probe-ban.conf -o /etc/fail2ban/filter.d/chronokeep-probe-ban.conf
curl -fL $BASE/jail.d/chronokeep-remote.conf -o /etc/fail2ban/jail.d/chronokeep-remote.conf
```

Before starting it, edit `/etc/fail2ban/jail.d/chronokeep-remote.conf`:

- **`ignoreip`**: add your own IPs and any networks in the API's `BAN_ALLOWLIST`, such as race
  venues or timing networks. Anything not listed can be blocked on every port, SSH included.
- **`logpath`**: the jail reads `/chronokeep/remote/logs/*.log`. If the service's output goes
  to the systemd journal instead, switch to the commented `journalmatch` lines.

Then:

```sh
systemctl enable --now fail2ban
systemctl reload fail2ban
```

fail2ban only looks for log files matching `logpath` when it starts or reloads. Whenever the
service starts a new log file (for example after a daily restart), reload the jail just after,
from root's crontab, or fail2ban keeps reading the previous file:

```
5 0 * * * fail2ban-client reload chronokeep-remote-probes >/dev/null 2>&1
```

Chronokeep Results has the same files with its own jail. Both use the same filter, so on a
server running both, copy the filter once and both jail files.

## Check it

See whether the filter finds ban lines in a log file:

```sh
fail2ban-regex /chronokeep/remote/logs/<file>.log /etc/fail2ban/filter.d/chronokeep-probe-ban.conf
```

or, if you switched to the journal:

```sh
fail2ban-regex systemd-journal[journalmatch='_SYSTEMD_UNIT=chronokeep-remote.service'] chronokeep-probe-ban
```

`fail2ban-client status chronokeep-remote-probes` lists the files the jail is reading, so you
can check it moved to the new file after a reload.

See what's banned, and unban an address:

```sh
fail2ban-client status chronokeep-remote-probes
fail2ban-client set chronokeep-remote-probes unbanip 203.0.113.5
```

An IP unbanned in fail2ban may still be banned in the API until its 24 hours run out or the
service restarts.
