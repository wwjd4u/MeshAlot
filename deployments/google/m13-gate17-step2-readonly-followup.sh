#!/usr/bin/env bash
# M13 Gate 17 Step 2 additional production baseline lookup: READ ONLY.
# Run via approved IAP SSH on meshalot-control-01. Output stays private on MS-02.
set -euo pipefail
test "$(hostname -s)" = "meshalot-control-01" || {
  echo "M13_FOLLOWUP=REFUSED_WRONG_HOST"
  exit 2
}
echo "M13_FOLLOWUP_HOST=VERIFIED"

echo "===== RUNNING MESHALOT SERVICE AND ACTUAL BINARY (NO ENVIRONMENT VALUES) ====="
found=0
while read -r unit enabled rest; do
  case "$unit" in
    meshalot*.service)
      found=$((found+1))
      echo "API_UNIT_NAME=$unit"
      state="$(systemctl show "$unit" -p ActiveState --value 2>/dev/null || true)"
      echo "API_UNIT_STATE=$state"
      mainpid="$(systemctl show "$unit" -p MainPID --value 2>/dev/null || true)"
      if [[ "$mainpid" =~ ^[0-9]+$ ]] && test "$mainpid" -gt 1; then
        echo "API_MAIN_PID_AVAILABLE=YES"
        binary="$(readlink -f "/proc/$mainpid/exe" 2>/dev/null || true)"
        if test -n "$binary"; then
          echo "API_RUNNING_EXE_PATH=$binary"
          if test -r "/proc/$mainpid/exe"; then
            digest="$(sha256sum "/proc/$mainpid/exe" 2>/dev/null | cut -d ' ' -f1 || true)"
          else
            digest=""
          fi
          if [[ "$digest" =~ ^[0-9a-f]{64}$ ]]; then
            echo "API_RUNNING_EXE_SHA256=$digest"
          else
            echo "API_RUNNING_EXE_SHA256=UNAVAILABLE"
          fi
        else
          echo "API_RUNNING_EXE_PATH=UNAVAILABLE"
        fi
      else
        echo "API_MAIN_PID_AVAILABLE=NO"
      fi
      ;;
  esac
done < <(systemctl list-unit-files 'meshalot*.service' --no-pager --no-legend 2>/dev/null || true)
echo "MESHALOT_UNIT_COUNT=$found"

echo "===== CADDY REVERSE PROXY LOOPBACK HEALTH (NO RAW CADDYFILE) ====="
CADDY=/etc/caddy/Caddyfile
LOOPBACKS=""
if test -r "$CADDY"; then
  LOOPBACKS="$(grep -oE '(127\.0\.0\.1|localhost|\[::1\]):[0-9]{2,5}' "$CADDY" 2>/dev/null | sort -u || true)"
elif command -v sudo >/dev/null 2>&1 && sudo -n test -r "$CADDY" 2>/dev/null; then
  LOOPBACKS="$(sudo -n grep -oE '(127\.0\.0\.1|localhost|\[::1\]):[0-9]{2,5}' "$CADDY" 2>/dev/null | sort -u || true)"
else
  echo "CADDY_FILE_LOOPBACK_QUERY=UNAVAILABLE"
fi
if test -z "$LOOPBACKS"; then
  echo "CADDY_LOOPBACK_ENDPOINTS=NONE_FOUND"
else
  echo "CADDY_LOOPBACK_ENDPOINTS=DETECTED"
  count=0
  while read -r target; do
    test -n "$target" || continue
    port="${target##*:}"
    if [[ "$port" =~ ^[0-9]{2,5}$ ]] && test "$port" -le 65535 && test "$port" -ge 1024; then
      if test "$count" -ge 6; then break; fi
      count=$((count+1))
      host="${target%:*}"
      if test "$host" = localhost; then host=127.0.0.1; fi
      if test "$host" = "[::1]"; then host="[::1]"; fi
      # Probe ONLY local loopback endpoints already referenced by Caddy.
      httpcode="$(curl --silent --show-error --output /dev/null \
        --connect-timeout 2 --max-time 4 --write-out '%{http_code}' \
        "http://$host:$port/v1/health" 2>/dev/null || true)"
      if test -z "$httpcode"; then httpcode=000; fi
      printf 'CADDY_PROXY_HEALTH_TARGET=%s HTTP_CODE=%s\n' "$target" "$httpcode"
    fi
  done <<<"$LOOPBACKS"
  echo "CADDY_LOOPBACK_ENDPOINTS_PROBED=$count"
fi

echo "M13_FOLLOWUP_READ_ONLY=PASS"
