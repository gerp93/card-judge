# Local MariaDB setup (for verifying schema + game flow)

How to stand up a throwaway MariaDB in a fresh Linux dev/sandbox container so you
can actually run the app, apply the schema, and play through a round. This is for
**local/ephemeral verification only** — not a production setup. Reproduced here so
future sessions don't have to rediscover it.

Prerequisites seen in the sandbox: `root`, `apt-get`, Go toolchain. No MariaDB,
no `sqlfmt`, no running service manager.

## 1. Install the server

```bash
apt-get update -qq
DEBIAN_FRONTEND=noninteractive apt-get install -y -qq mariadb-server
```

The Ubuntu package initializes the data directory at `/var/lib/mysql` during
install. It does **not** start the service (containers deny `invoke-rc.d` /
`policy-rc.d`), so start it by hand next. (Unrelated `deadsnakes`/`php` PPA 403s
from the proxy during `apt-get update` are harmless — the ubuntu archive still
resolves.)

## 2. Start it manually

```bash
mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
nohup mariadbd-safe > /tmp/mariadb.log 2>&1 &
sleep 8
mysqladmin ping        # -> "mysqld is alive"
```

Default bind is `127.0.0.1:3306`, which is what the app's TCP DSN expects. `root`
authenticates over the unix socket with no password (`mysql -u root`).

## 3. Create the app user + database

The app connects over TCP as a normal user, so create one (the schema uses
stored procedures/functions/triggers/events, hence the broad grant):

```bash
mysql -u root <<'SQL'
CREATE USER IF NOT EXISTS 'carduser'@'127.0.0.1' IDENTIFIED BY 'cardpass';
CREATE USER IF NOT EXISTS 'carduser'@'localhost' IDENTIFIED BY 'cardpass';
GRANT ALL PRIVILEGES ON *.* TO 'carduser'@'127.0.0.1' WITH GRANT OPTION;
GRANT ALL PRIVILEGES ON *.* TO 'carduser'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
SQL

# setup.sql DROPs and re-CREATEs the CARD_JUDGE database (it is NOT in the
# auto-run SQLFiles manifest, so run it once by hand):
mysql -u root < src/static/sql/setup.sql
```

## 4. Build + run (applies the rest of the schema on startup)

The server iterates `static.SQLFiles` and applies every remaining SQL object on
boot, so just running it loads the schema.

```bash
cd src && go build -o /tmp/card-judge .
export CARD_JUDGE_SQL_HOST=127.0.0.1 \
       CARD_JUDGE_SQL_DATABASE=CARD_JUDGE \
       CARD_JUDGE_SQL_USER=carduser \
       CARD_JUDGE_SQL_PASSWORD=cardpass \
       CARD_JUDGE_PORT=2016
/tmp/card-judge &            # logs "server is running..."
curl -s -o /dev/null -w "HTTP %{http_code}\n" http://127.0.0.1:2016/login   # -> HTTP 200
```

## 5. Sanity-check the schema loaded

```bash
mysql -u root -e "SELECT COUNT(*) FROM information_schema.tables    WHERE table_schema='CARD_JUDGE';"   # tables+views (29 on a clean checkout)
mysql -u root -e "SELECT COUNT(*) FROM information_schema.routines  WHERE routine_schema='CARD_JUDGE';" # procedures+functions (58 on a clean checkout)
```

Baseline counts on a clean `main`: **29 tables/views, 58 routines.** After the
Gameshell base/extension table split these numbers change (two new `CJ_*`
extension tables); re-baseline against your branch, don't assume the old numbers.

## Re-applying schema during development

`static.SQLFiles` uses `CREATE OR REPLACE` (routines/views/triggers) and
`CREATE TABLE IF NOT EXISTS` (tables), so restarting the app re-applies changed
objects. For a clean slate (e.g. after changing a table definition), re-run
`setup.sql` to drop+recreate the database, then start the app again.

## Notes / gotchas

- No `sqlfmt` binary in the sandbox — the repo's `src/static/sql/sqlfmt.sh` can't
  run here. Match the existing SQL formatting (uppercase keywords + identifiers,
  4-space indent) by hand instead.
- `mariadbd-safe` runs in the background for the life of the container; it is not
  a managed service and won't survive a container reclaim. Re-run steps 2–4 in a
  new session.
- This is ephemeral: data is thrown away with the container. Never point this at
  anything real.
