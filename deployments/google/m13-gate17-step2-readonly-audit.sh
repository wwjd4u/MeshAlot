#!/usr/bin/env bash
# M13 Gate 17 Step 2 — READ-ONLY production control-server audit.
# Run only on meshalot-control-01 via the approved Google Cloud IAP route.
# Never print DSNs, passwords, private keys, raw Caddyfile or systemd Environment.
set -euo pipefail
umask 077

server_name="$(hostname -s)"
if test "$server_name" != "meshalot-control-01"; then
  echo "M13_GATE17_STEP2=REFUSED_WRONG_HOST"
  exit 2
fi
echo "M13_GATE17_STEP2_HOSTNAME_CONFIRMED=PASS"
printf 'AUDIT_UTC=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"

echo "===== API SERVICE NAMES AND STATES (NO ENVIRONMENT OUTPUT) ====="
if command -v systemctl >/dev/null 2>&1; then
  while read -r unit enabled rest; do
    case "$unit" in
      meshalot*.service)
        echo "UNIT_NAME=$unit ENABLED_STATE=$enabled"
        for field in LoadState ActiveState FragmentPath DropInPaths; do
          # Allowlist excludes Environment and ExecStart argument credentials.
          value="$(systemctl show "$unit" --property="$field" --value 2>/dev/null || true)"
          if test -z "$value"; then value="UNAVAILABLE"; fi
          printf 'UNIT_%s_%s=%s\n' "$unit" "$field" "$value"
        done
        ;;
    esac
  done < <(systemctl list-unit-files 'meshalot*.service' --no-pager --no-legend 2>/dev/null || true)
fi
echo "M13_GATE17_STEP2_SYSTEMD_METADATA=READ_ONLY"

echo "===== ACTIVE RELEASE LINKS AND BINARY FINGERPRINTS ====="
for link in /opt/meshalot/current /var/www/meshalot/current; do
  if test -L "$link"; then
    target="$(readlink "$link")"
    printf 'RELEASE_LINK=%s TARGET=%s\n' "$link" "$target"
  fi
done
for binary in /opt/meshalot/current/meshalot-server /var/www/meshalot/current/meshalot-server; do
  if test -f "$binary" && test -r "$binary"; then
    printf 'ACTIVE_BINARY=%s ' "$binary"
    sha256sum "$binary" | cut -d ' ' -f 1
  fi
done
if test -d /opt/meshalot/source/.git; then
  source_sha="$(git -C /opt/meshalot/source rev-parse HEAD 2>/dev/null || true)"
  if [[ "$source_sha" =~ ^[0-9a-f]{40}$ ]]; then
    echo "PRODUCTION_SOURCE_HEAD=$source_sha"
  else
    echo "PRODUCTION_SOURCE_HEAD=UNVERIFIED"
  fi
fi

echo "===== CADDY DIRECTIVE PRESENCE (CONFIG CONTENT NOT PRINTED) ====="
caddy_path="/etc/caddy/Caddyfile"
caddy_read="none"
if test -r "$caddy_path"; then
  caddy_read="direct"
elif command -v sudo >/dev/null 2>&1 && sudo -n test -r "$caddy_path" 2>/dev/null; then
  caddy_read="sudo"
fi
if test "$caddy_read" = "none"; then
  echo "CADDY_CONFIG=UNREADABLE"
else
  caddy_grep() {
    if test "$caddy_read" = "sudo"; then
      sudo -n grep -Eq -- "$1" "$caddy_path" 2>/dev/null
    else
      grep -Eq -- "$1" "$caddy_path" 2>/dev/null
    fi
  }
  if caddy_grep 'reverse_proxy'; then
    echo "CADDY_REVERSE_PROXY_DIRECTIVE=PRESENT"
  else
    echo "CADDY_REVERSE_PROXY_DIRECTIVE=NOT_FOUND"
  fi
  if caddy_grep '/v1/|/v1\*|/v1/\*'; then
    echo "CADDY_API_PATH_REFERENCE=PRESENT"
  else
    echo "CADDY_API_PATH_REFERENCE=NOT_FOUND"
  fi
  if caddy_grep '127\.0\.0\.1|localhost|::1'; then
    echo "CADDY_LOOPBACK_UPSTREAM_REFERENCE=PRESENT"
  else
    echo "CADDY_LOOPBACK_UPSTREAM_REFERENCE=NOT_FOUND"
  fi
  echo "CADDY_DIRECTIVE_AUDIT=PRELIMINARY_REQUIRES_OPERATOR_REVIEW"
fi

echo "===== TRACKED POSTGRES MIGRATIONS (READ ONLY; NO USER DATA) ====="
if command -v sudo >/dev/null 2>&1 && command -v psql >/dev/null 2>&1; then
  if sudo -n -u postgres psql -X -qAt -v ON_ERROR_STOP=1 \
     --dbname=meshalot --no-password 2>/dev/null <<'SQL'
BEGIN READ ONLY;
SELECT 'MIGRATION=' || version::text || ':' || filename || ':' || left(sha256,12)
FROM meshalot_meta.schema_migrations ORDER BY version;
ROLLBACK;
SQL
  then
    echo "M13_GATE17_STEP2_MIGRATION_QUERY=PASS_READ_ONLY"
  else
    echo "M13_GATE17_STEP2_MIGRATION_QUERY=UNAVAILABLE_NO_CHANGES"
  fi
else
  echo "M13_GATE17_STEP2_MIGRATION_QUERY=UNAVAILABLE_NO_CHANGES"
fi

echo "===== EXISTING PRIVATE BACKUP INVENTORY (NAMES ONLY) ====="
backups="/opt/meshalot/backups"
if test -d "$backups" && test -r "$backups"; then
  find "$backups" -mindepth 1 -maxdepth 1 -type d -printf 'BACKUP_DIR=%f\n' 2>/dev/null | head -35 || true
  echo "BACKUP_INVENTORY=LISTED_READ_ONLY"
elif command -v sudo >/dev/null 2>&1 && sudo -n test -d "$backups" 2>/dev/null; then
  sudo -n find "$backups" -mindepth 1 -maxdepth 1 -type d -printf 'BACKUP_DIR=%f\n' 2>/dev/null | head -35 || true
  echo "BACKUP_INVENTORY=LISTED_READ_ONLY"
else
  echo "BACKUP_INVENTORY=UNAVAILABLE"
fi

echo "===== SAFE HEALTH PROBES ====="
if command -v curl >/dev/null 2>&1; then
  status="$(curl --silent --show-error --output /dev/null --max-time 5 \
    --write-out '%{http_code}' 'http://127.0.0.1:8080/v1/health' 2>/dev/null || true)"
  if test -z "$status"; then status="UNAVAILABLE"; fi
  printf 'LOCAL_BACKEND_HEALTH_HTTP=%s\n' "$status"
fi
echo "M13_GATE17_STEP2_READ_ONLY_AUDIT_COMPLETE=YES"
echo "M13_GATE17_STEP2_PRODUCTION_CHANGES=NONE"
