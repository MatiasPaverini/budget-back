#!/usr/bin/env bash
# Seeds the local dev database with a realistic set of accounts, transactions,
# and IPC history, so the frontend has something to look at without you
# having to type it all in by hand. Safe to re-run: it wipes the existing
# accounts for the dev user first (cascades their transactions), and the
# IPC points are upserted by period so re-running just overwrites them.
#
# Requires: the backend running with AUTH_MODE=dev (no token needed), curl, jq.
set -euo pipefail

API_URL="${API_URL:-http://localhost:8081}"

echo "Seeding against $API_URL ..."

# --- wipe existing accounts (cascades transactions) ---
existing_ids=$(curl -s "$API_URL/accounts" | jq -r '.[].id')
for id in $existing_ids; do
  curl -s -o /dev/null -X DELETE "$API_URL/accounts/$id"
done

create_account() {
  curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d "$1" | jq -r '.id'
}

create_transaction() {
  local account_id="$1" amount="$2" description="$3" category="${4:-}"
  local category_json="null"
  if [ -n "$category" ]; then
    category_json="\"$category\""
  fi
  curl -s -o /dev/null -X POST "$API_URL/transactions" -H "Content-Type: application/json" -d "{
    \"account_id\": \"$account_id\",
    \"amount\": \"$amount\",
    \"description\": \"$description\",
    \"category\": $category_json
  }"
}

record_ipc() {
  curl -s -o /dev/null -X POST "$API_URL/indicators/ipc" -H "Content-Type: application/json" -d "{
    \"period\": \"$1\",
    \"value\": \"$2\"
  }"
}

echo "Creating accounts..."

CASH=$(create_account '{"name":"Efectivo","type":"cash","currency":"ARS","opening_balance":"5000"}')
BANK=$(create_account '{"name":"Cuenta Corriente","type":"bank","currency":"ARS","opening_balance":"50000"}')
CARD=$(create_account '{"name":"Tarjeta Visa","type":"credit_card","currency":"ARS","opening_balance":"0","credit_limit":"500000","statement_close_day":25,"due_day":10}')
LINE=$(create_account '{"name":"Adelanto en Cuenta","type":"credit_line","currency":"ARS","opening_balance":"0","credit_limit":"100000"}')
INVEST=$(create_account '{"name":"Plazo Fijo","type":"investment","currency":"ARS","opening_balance":"300000"}')
LOAN=$(create_account '{"name":"Préstamo Personal","type":"loan","currency":"ARS","opening_balance":"-400000","interest_rate":"45.5","term_months":24}')

echo "Adding transactions..."

create_transaction "$CASH" "-3000" "Verdulería" "groceries"
create_transaction "$CASH" "10000" "Regalo cumpleaños" "gift"

create_transaction "$BANK" "450000" "Sueldo" "salary"
create_transaction "$BANK" "-85000" "Alquiler" "rent"
create_transaction "$BANK" "-30000" "Terapia" "health"
create_transaction "$BANK" "-15000" "Supermercado" "groceries"
create_transaction "$BANK" "-8000" "Netflix + Spotify" "subscriptions"

create_transaction "$CARD" "-45000" "Compras varias" "shopping"
create_transaction "$CARD" "-22000" "Nafta" "transport"
create_transaction "$CARD" "45000" "Pago tarjeta" "payment"

create_transaction "$LINE" "-30000" "Adelanto en efectivo" "cash-advance"

create_transaction "$INVEST" "12000" "Interés plazo fijo" "interest"

create_transaction "$LOAN" "20000" "Cuota 1" "loan-payment"
create_transaction "$LOAN" "20000" "Cuota 2" "loan-payment"

echo "Recording IPC history..."

record_ipc "2026-01-01T00:00:00Z" "100.0"
record_ipc "2026-02-01T00:00:00Z" "104.5"
record_ipc "2026-03-01T00:00:00Z" "109.2"
record_ipc "2026-04-01T00:00:00Z" "113.8"
record_ipc "2026-05-01T00:00:00Z" "118.9"
record_ipc "2026-06-01T00:00:00Z" "124.1"

echo "Done. Net worth:"
curl -s "$API_URL/networth" | jq
