#!/usr/bin/env bash
# End-to-end tests for the goth server, driven through the gothtest CLI.
#
# Needs a running server (with Postgres + Redis) whose mail goes to mailpit,
# so OTP codes can be read from the emails like a real user would.
#
#   GOTHTEST     path to the gothtest binary          (default: client/cli/gothtest)
#   TOTPCODE     path to the totpcode binary          (default: go run ./scripts/e2e/totpcode)
#   GOTH_ADDR    gRPC server address                  (default: localhost:50051)
#   MAILPIT_URL  mailpit HTTP API                     (default: http://localhost:8025)
#   EMAIL_DOMAIN domain for generated test emails     (default: example.com)

set -uo pipefail

GOTHTEST=${GOTHTEST:-client/cli/gothtest}
GOTH_ADDR=${GOTH_ADDR:-localhost:50051}
MAILPIT_URL=${MAILPIT_URL:-http://localhost:8025}
EMAIL_DOMAIN=${EMAIL_DOMAIN:-example.com}
RUN_ID=${RUN_ID:-$(date +%s)$RANDOM}
PASSWORD="e2e-pass-1"

PASS=0
FAIL=0
OUT=""
declare -A LAST_MAIL_ID

# ---------- helpers ----------

totp_code() {
	if [[ -n "${TOTPCODE:-}" ]]; then "$TOTPCODE" "$1"; else go run ./scripts/e2e/totpcode "$1"; fi
}

# run <gothtest args...>: runs the CLI, output in $OUT, exit status in $RC
run() {
	OUT=$("$GOTHTEST" --addr "$GOTH_ADDR" "$@" 2>&1)
	RC=$?
}

pass() { PASS=$((PASS + 1)); printf '  \033[32m✔\033[0m %s\n' "$1"; }
fail() {
	FAIL=$((FAIL + 1))
	printf '  \033[31m✘ %s\033[0m\n' "$1"
	printf '%s\n' "$OUT" | grep -v -E '^\s+--|^Usage:|^Flags:|^Global Flags:|^\s*gothtest |^$' | sed 's/^/      /' | head -8
}

# expect_ok <name> <gothtest args...>: the call must succeed
expect_ok() {
	local name=$1; shift
	run "$@"
	if [[ $RC -eq 0 ]]; then pass "$name"; else fail "$name (expected success)"; fi
}

# expect_err <name> <grpc code> <gothtest args...>: the call must fail with that code
expect_err() {
	local name=$1 code=$2; shift 2
	run "$@"
	if [[ $RC -ne 0 && "$OUT" == *"code = $code"* ]]; then pass "$name → $code"; else fail "$name (expected $code)"; fi
}

# expect_out <name> <text>: the last output must contain text
expect_out() {
	if [[ "$OUT" == *"$2"* ]]; then pass "$1"; else fail "$1 (expected output: $2)"; fi
}

# field <label>: value printed after "<label>:" in the last output
field() {
	printf '%s\n' "$OUT" | sed -n "s/^$1: *//p" | head -1
}

# check <name> <actual> <expected>: plain value comparison
check() {
	if [[ "$2" == "$3" ]]; then pass "$1"; else OUT="got: $2, want: $3"; fail "$1"; fi
}

newest_mail_id() {
	curl -s --get "$MAILPIT_URL/api/v1/search" --data-urlencode "query=to:\"$1\"" --data-urlencode "limit=1" \
		| jq -r '.messages[0].ID // empty'
}

# mail_code <email>: waits for a new email to that address and sets MAIL_CODE to its 6-digit code.
# Not called inside $(...): a subshell would lose the LAST_MAIL_ID update.
mail_code() {
	local to=$1 id last=${LAST_MAIL_ID[$1]:-} i
	MAIL_CODE="no-mail"
	for i in $(seq 1 30); do
		id=$(newest_mail_id "$to")
		if [[ -n "$id" && "$id" != "$last" ]]; then
			LAST_MAIL_ID[$to]=$id
			MAIL_CODE=$(curl -s "$MAILPIT_URL/api/v1/message/$id" | jq -r '.HTML' | grep -oP '<h1>\K[0-9]{6}' | head -1)
			return 0
		fi
		sleep 0.5
	done
	return 1
}

# expect_no_new_mail <name> <email>: no email arrives for that address
expect_no_new_mail() {
	sleep 2
	local id
	id=$(newest_mail_id "$2")
	if [[ "$id" == "${LAST_MAIL_ID[$2]:-}" ]]; then pass "$1"; else OUT="a new email arrived"; fail "$1"; fi
}

email_for() { echo "e2e-$RUN_ID-$1@$EMAIL_DOMAIN"; }

# new_user <name>: signs up, verifies the email with the mailed code and logs in.
# Sets <NAME>_EMAIL, <NAME>_AT (access token), <NAME>_RT (refresh token).
new_user() {
	local name=$1 email code
	email=$(email_for "$name")
	run signup --email "$email" --name "E2E $name" --password "$PASSWORD"
	[[ $RC -eq 0 ]] || { fail "setup: signup $name"; return 1; }
	mail_code "$email"; code=$MAIL_CODE
	run verify-otp --purpose 2 --email "$email" --code "$code"
	[[ $RC -eq 0 ]] || { fail "setup: verify $name"; return 1; }
	run login --email "$email" --password "$PASSWORD"
	[[ $RC -eq 0 ]] || { fail "setup: login $name"; return 1; }
	printf -v "${name^^}_EMAIL" '%s' "$email"
	printf -v "${name^^}_AT" '%s' "$(field 'Access Token')"
	printf -v "${name^^}_RT" '%s' "$(field 'Refresh Token')"
}

section() { printf '\n\033[1m%s\033[0m\n' "$1"; }

# ---------- signup + email verification ----------

section "Signup and email verification"
EMAIL=$(email_for signup)
expect_err "signup with missing password" InvalidArgument signup --email "$EMAIL" --name "E2E" --password ""
expect_ok  "signup" signup --email "$EMAIL" --name "E2E" --password "$PASSWORD"
expect_out "signup returns email_verified=false" "Email Verified: false"
mail_code "$EMAIL"; SIGNUP_CODE=$MAIL_CODE
check "signup emails a verification code" "${#SIGNUP_CODE}" 6
expect_err "login before verifying" PermissionDenied login --email "$EMAIL" --password "$PASSWORD"
expect_err "resend verification within cooldown" ResourceExhausted send-otp --purpose 2 --email "$EMAIL"
expect_err "verify-otp without email" InvalidArgument verify-otp --purpose 2 --code "$SIGNUP_CODE"
expect_err "verify-otp with wrong code" PermissionDenied verify-otp --purpose 2 --email "$EMAIL" --code 000000
expect_ok  "verify-otp with mailed code" verify-otp --purpose 2 --email "$EMAIL" --code "$SIGNUP_CODE"
expect_err "verify-otp reusing the code" NotFound verify-otp --purpose 2 --email "$EMAIL" --code "$SIGNUP_CODE"
expect_err "send verification to verified account" AlreadyExists send-otp --purpose 2 --email "$EMAIL"
expect_err "verify-otp with MFA purpose" InvalidArgument verify-otp --purpose 0 --email "$EMAIL" --code 123456
expect_err "verify-otp with unknown purpose" InvalidArgument verify-otp --purpose 9 --email "$EMAIL" --code 123456

# ---------- login ----------

section "Login"
expect_err "login with missing password" InvalidArgument login --email "$EMAIL" --password ""
expect_err "login with unknown email" Unauthenticated login --email "$(email_for nobody)" --password "$PASSWORD"
expect_err "login with wrong password" Unauthenticated login --email "$EMAIL" --password wrong-pass
expect_ok  "login after verifying" login --email "$EMAIL" --password "$PASSWORD"
expect_out "login without MFA returns tokens" "Access Token:"

# ---------- tokens ----------

section "Access and refresh tokens"
new_user tok
expect_ok  "validate-token with a valid token" validate-token --access-token "$TOK_AT"
expect_out "validate-token says valid" "Valid:        true"
expect_out "validate-token returns the email" "$TOK_EMAIL"
expect_ok  "validate-token with garbage" validate-token --access-token garbage
expect_out "garbage token is not valid" "Valid:        false"
expect_ok  "get-user own profile" get-user --access-token "$TOK_AT"
expect_out "get-user returns the email" "$TOK_EMAIL"
expect_err "get-user another user's id" PermissionDenied get-user --access-token "$TOK_AT" --user-id 00000000-0000-0000-0000-000000000000
expect_err "get-user with garbage token" Unauthenticated get-user --access-token garbage
expect_ok  "refresh-token without access token" refresh-token --refresh-token "$TOK_RT"
NEW_RT=$(field 'Refresh Token')
expect_err "refresh-token with the rotated-out token" Unauthenticated refresh-token --refresh-token "$TOK_RT"
expect_err "refresh-token with garbage" Unauthenticated refresh-token --refresh-token garbage
expect_ok  "logout without access token" logout --refresh-token "$NEW_RT"
expect_err "refresh-token after logout" Unauthenticated refresh-token --refresh-token "$NEW_RT"
expect_err "logout with garbage" Unauthenticated logout --refresh-token garbage

# ---------- send-otp auth ----------

section "send-otp (MFA purpose) auth"
expect_err "MFA otp with no token" Unauthenticated send-otp --purpose 0
expect_err "MFA otp with garbage access token" Unauthenticated send-otp --purpose 0 --access-token garbage
expect_err "MFA otp with garbage session token" Unauthenticated send-otp --purpose 0 --mfa-session-token garbage
expect_ok  "MFA otp with access token" send-otp --purpose 0 --access-token "$TOK_AT" --email "$(email_for someone-else)"
mail_code "$TOK_EMAIL"; CODE=$MAIL_CODE
check "MFA otp goes to the token's email, not --email" "${#CODE}" 6
expect_err "MFA otp resend within cooldown" ResourceExhausted send-otp --purpose 0 --access-token "$TOK_AT"
expect_err "send-otp with unknown purpose" InvalidArgument send-otp --purpose 9 --email "$TOK_EMAIL"

# ---------- password reset ----------

section "Password reset codes"
new_user reset
expect_ok  "send reset code" send-otp --purpose 1 --email "$RESET_EMAIL"
mail_code "$RESET_EMAIL"; RESET_CODE=$MAIL_CODE
expect_err "resend reset code within cooldown" ResourceExhausted send-otp --purpose 1 --email "$RESET_EMAIL"
expect_err "verify reset without email" InvalidArgument verify-otp --purpose 1 --code "$RESET_CODE"
expect_err "verify reset with wrong code" PermissionDenied verify-otp --purpose 1 --email "$RESET_EMAIL" --code 000000
expect_ok  "verify reset with mailed code" verify-otp --purpose 1 --email "$RESET_EMAIL" --code "$RESET_CODE"
expect_out "reset returns a session token" "Session Token:"
expect_err "verify reset reusing the code" NotFound verify-otp --purpose 1 --email "$RESET_EMAIL" --code "$RESET_CODE"

new_user burn
expect_ok  "send reset code (burn test)" send-otp --purpose 1 --email "$BURN_EMAIL"
mail_code "$BURN_EMAIL"; BURN_CODE=$MAIL_CODE
for i in 1 2 3 4; do
	expect_err "wrong reset code #$i" PermissionDenied verify-otp --purpose 1 --email "$BURN_EMAIL" --code 000000
done
expect_err "wrong reset code #5 burns the code" ResourceExhausted verify-otp --purpose 1 --email "$BURN_EMAIL" --code 000000
expect_err "correct code after burn" NotFound verify-otp --purpose 1 --email "$BURN_EMAIL" --code "$BURN_CODE"

# ---------- update-password ----------

section "Update password"
new_user pw
expect_err "wrong old password" PermissionDenied update-password --access-token "$PW_AT" --old-password wrong-pass --new-password new-pass-1
expect_err "new password same as old" InvalidArgument update-password --access-token "$PW_AT" --old-password "$PASSWORD" --new-password "$PASSWORD"
expect_err "another user's id" PermissionDenied update-password --access-token "$PW_AT" --user-id 00000000-0000-0000-0000-000000000000 --old-password "$PASSWORD" --new-password new-pass-1
expect_ok  "change password" update-password --access-token "$PW_AT" --old-password "$PASSWORD" --new-password new-pass-1
expect_err "old refresh token is revoked" Unauthenticated refresh-token --refresh-token "$PW_RT"
expect_err "login with old password" Unauthenticated login --email "$PW_EMAIL" --password "$PASSWORD"
expect_ok  "login with new password" login --email "$PW_EMAIL" --password new-pass-1

# ---------- enable-mfa ----------

section "Enable MFA"
new_user totp
expect_err "enable unknown MFA type" InvalidArgument enable-mfa --access-token "$TOTP_AT" --mfa-type 5
expect_err "enable MFA type none" InvalidArgument enable-mfa --access-token "$TOTP_AT" --mfa-type 0
expect_err "enable MFA with garbage token" Unauthenticated enable-mfa --access-token garbage --mfa-type 1
expect_ok  "enable TOTP" enable-mfa --access-token "$TOTP_AT" --mfa-type 1
TOTP_SECRET=$(field 'TOTP Secret')
expect_out "enable TOTP returns a QR url" "otpauth://"
expect_err "enable MFA twice" AlreadyExists enable-mfa --access-token "$TOTP_AT" --mfa-type 2

# ---------- TOTP login ----------

section "TOTP login"
expect_ok  "login with TOTP on" login --email "$TOTP_EMAIL" --password "$PASSWORD"
expect_out "login says MFA required" "MFA Required"
SESSION=$(field 'MFA Session Token')
if [[ "$OUT" != *"Access Token"* ]]; then pass "login with MFA on returns no tokens"; else fail "login with MFA on returns no tokens"; fi
expect_err "verify-mfa with garbage session" PermissionDenied verify-mfa --mfa-session-token garbage --code 123456
expect_err "verify-mfa with wrong code" PermissionDenied verify-mfa --mfa-session-token "$SESSION" --code 000000
expect_ok  "verify-mfa with TOTP code" verify-mfa --mfa-session-token "$SESSION" --code "$(totp_code "$TOTP_SECRET")"
MFA_AT=$(field 'Access Token')
MFA_RT=$(field 'Refresh Token')
expect_ok  "refresh-token from an MFA login" refresh-token --refresh-token "$MFA_RT"
expect_ok  "validate-token from an MFA login" validate-token --access-token "$MFA_AT"
expect_out "MFA token is mfa_verified" "MFA Verified: true"

# ---------- disable TOTP ----------

section "Disable TOTP"
expect_err "disable with wrong code" PermissionDenied disable-mfa --access-token "$MFA_AT" --code 000000
expect_ok  "disable with TOTP code" disable-mfa --access-token "$MFA_AT" --code "$(totp_code "$TOTP_SECRET")"
expect_err "disable again" FailedPrecondition disable-mfa --access-token "$MFA_AT" --code 123456
expect_err "verify-mfa after disable" FailedPrecondition verify-mfa --mfa-session-token "$SESSION" --code 123456
expect_ok  "login after disable" login --email "$TOTP_EMAIL" --password "$PASSWORD"
expect_out "login after disable returns tokens" "Access Token:"

# ---------- TOTP lockout ----------

section "TOTP lockout"
new_user lock
expect_ok  "enable TOTP" enable-mfa --access-token "$LOCK_AT" --mfa-type 1
LOCK_SECRET=$(field 'TOTP Secret')
expect_ok  "login" login --email "$LOCK_EMAIL" --password "$PASSWORD"
LOCK_SESSION=$(field 'MFA Session Token')
for i in 1 2 3 4; do
	expect_err "wrong TOTP #$i" PermissionDenied verify-mfa --mfa-session-token "$LOCK_SESSION" --code 000000
done
expect_err "wrong TOTP #5 locks the user" ResourceExhausted verify-mfa --mfa-session-token "$LOCK_SESSION" --code 000000
expect_err "correct TOTP while locked" ResourceExhausted verify-mfa --mfa-session-token "$LOCK_SESSION" --code "$(totp_code "$LOCK_SECRET")"
expect_err "disable while locked" ResourceExhausted disable-mfa --access-token "$LOCK_AT" --code "$(totp_code "$LOCK_SECRET")"

# ---------- email OTP login ----------

section "Email OTP login"
new_user mail
expect_ok  "enable email OTP" enable-mfa --access-token "$MAIL_AT" --mfa-type 2
mail_code "$MAIL_EMAIL"
expect_ok  "login with email OTP on" login --email "$MAIL_EMAIL" --password "$PASSWORD"
expect_out "login says MFA required" "MFA Required"
MAIL_SESSION=$(field 'MFA Session Token')
mail_code "$MAIL_EMAIL"; LOGIN_CODE=$MAIL_CODE
check "login emails an MFA code" "${#LOGIN_CODE}" 6
expect_err "verify-mfa with wrong code" PermissionDenied verify-mfa --mfa-session-token "$MAIL_SESSION" --code 000000
expect_ok  "verify-mfa with mailed code" verify-mfa --mfa-session-token "$MAIL_SESSION" --code "$LOGIN_CODE"
expect_err "verify-mfa reusing the code" NotFound verify-mfa --mfa-session-token "$MAIL_SESSION" --code "$LOGIN_CODE"
expect_ok  "login again within cooldown" login --email "$MAIL_EMAIL" --password "$PASSWORD"
expect_no_new_mail "no new code sent within cooldown" "$MAIL_EMAIL"
expect_err "resend MFA code within cooldown" ResourceExhausted send-otp --purpose 0 --mfa-session-token "$MAIL_SESSION"

# ---------- email OTP burn ----------

section "Email OTP burn"
new_user mburn
expect_ok  "enable email OTP" enable-mfa --access-token "$MBURN_AT" --mfa-type 2
mail_code "$MBURN_EMAIL"
expect_ok  "login" login --email "$MBURN_EMAIL" --password "$PASSWORD"
MBURN_SESSION=$(field 'MFA Session Token')
mail_code "$MBURN_EMAIL"; MBURN_CODE=$MAIL_CODE
for i in 1 2 3 4; do
	expect_err "wrong MFA code #$i" PermissionDenied verify-mfa --mfa-session-token "$MBURN_SESSION" --code 000000
done
expect_err "wrong MFA code #5 burns the code" ResourceExhausted verify-mfa --mfa-session-token "$MBURN_SESSION" --code 000000
expect_err "correct code after burn" NotFound verify-mfa --mfa-session-token "$MBURN_SESSION" --code "$MBURN_CODE"

# ---------- disable email OTP ----------

section "Disable email OTP"
new_user mdis
expect_ok  "enable email OTP" enable-mfa --access-token "$MDIS_AT" --mfa-type 2
mail_code "$MDIS_EMAIL"; MDIS_CODE=$MAIL_CODE
expect_err "disable with wrong code" PermissionDenied disable-mfa --access-token "$MDIS_AT" --code 000000
expect_ok  "disable with mailed code" disable-mfa --access-token "$MDIS_AT" --code "$MDIS_CODE"
expect_err "disable again" FailedPrecondition disable-mfa --access-token "$MDIS_AT" --code "$MDIS_CODE"
expect_ok  "login after disable" login --email "$MDIS_EMAIL" --password "$PASSWORD"
expect_out "login after disable returns tokens" "Access Token:"

# ---------- summary ----------

printf '\n\033[1m%d passed, %d failed\033[0m\n' "$PASS" "$FAIL"
[[ $FAIL -eq 0 ]]
