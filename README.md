# goth

A gRPC authentication microservice written in Go. It provides email/password signup and login, JWT access and refresh tokens with server-side refresh-token rotation, and multi-factor authentication via TOTP or emailed one-time codes.

Messages are serialized with [Apache Fory](https://github.com/apache/fory) instead of Protocol Buffers. The service contract lives in `schema/goth.fdl`, and the Go types, gRPC stubs, and codec in `grpc/goth/` are generated from it by the Fory compiler.

> **Status:** Raft consensus for token revocation across goth nodes is being implemented. See [PR #61](https://github.com/ayush00git/goth/pull/61).

## Architecture

```
client ──gRPC/HTTP2 (Fory codec)──▶ :50051
                                      │
                    ┌─────────────────┼─────────────────┐
                    │ Recovery → Logging → Auth         │  unary interceptor chain
                    └─────────────────┼─────────────────┘
                                      ▼
                              GothService handlers
                                 │           │
                          PostgreSQL       Redis
                        (users, tokens,  (short-lived
                         MFA secrets)     OTP hashes)
                                 │
                            SMTP (Gmail)
```

- **Transport:** gRPC over TCP on port `50051`. The server forces the Fory `CodecV2` for every request, so a stock protobuf client cannot talk to this server.
- **Interceptors** (`internal/interceptors/`), applied in order:
  - `RecoveryInterceptor` converts panics into `codes.Internal` and logs the stack.
  - `LoggingInterceptor` logs method name, duration, and error or response type for every call.
  - `AuthInterceptor` requires an `authorization: Bearer <access_token>` metadata entry on every RPC except `Signup`, `Login`, `SendOTP`, and `VerifyOTP`. Validated claims are placed on the request context.
- **Handlers** (`internal/handlers/`) implement the RPCs against a `pgxpool.Pool` and a `go-redis` client.
- **Services** (`internal/services/`) hold token generation and validation, SMTP delivery, and TOTP URL construction.

## Requirements

- Go 1.25 or newer
- PostgreSQL with the `pgcrypto` extension
- Redis at `localhost:6379`
- An SMTP account on `smtp.gmail.com:587` for email OTP delivery

## Configuration

The server loads a `.env` file from the working directory and reads the following variables:

| Variable | Purpose |
|---|---|
| `DB_CONN_URI` | PostgreSQL connection string |
| `ACCESS_TOKEN_SECRET` | HS256 signing key for access tokens |
| `REFRESH_TOKEN_SECRET` | HS256 signing key for refresh tokens |
| `MFA_SESSION_TOKEN_SECRET` | HS256 signing key for the short-lived MFA session token |
| `MAIL_USERNAME` | SMTP username, also used as the `From` address |
| `MAIL_PASS` | SMTP password |

## Running the server

1. Apply the schema:

   ```sh
   psql "$DB_CONN_URI" -f migrations/001_init.sql
   ```

2. Start Redis on `localhost:6379`.

3. Create `.env` and fill in every variable listed above.

4. Build and run:

   ```sh
   go build ./...
   go run ./cmd/server
   ```

   On success the log shows the PostgreSQL connection, a Redis `PONG`, and `Goth server running on port :50051`.

## Test client

`client/cli/` is a separate Go module containing `gothtest`, a Cobra/Viper command-line client that speaks the Fory codec.

```sh
cd client/cli
make build          # produces ./gothtest
./gothtest --help
```

| Command | Flags |
|---|---|
| `signup` | `--email`, `--name`, `--password`, `--random-email` |
| `login` | `--email`, `--password`, `--device` |
| `logout` | `--refresh-token` |
| `refresh-token` | `--refresh-token` |
| `enable-mfa` | `--access-token`, `--mfa-type` (0 none, 1 TOTP, 2 email OTP) |
| `disable-mfa` | `--access-token`, `--code` |

The server address defaults to `localhost:50051` and can be changed with `--addr`, the `GOTHTEST_ADDR` environment variable, or a `$HOME/.gothtest.yaml` config file.

## Authentication model

**Passwords** are hashed with bcrypt before storage.

**Tokens** are JWTs signed with HS256. Each carries `user_id`, `email`, and standard registered claims.

| Token | Lifetime | Notes |
|---|---|---|
| Access token | 15 minutes | Presented as `authorization: Bearer <token>` metadata. Validated by signature and expiry. |
| Refresh token | 24 hours | Carries a `jti` and device fingerprint. The `jti` is stored in the `refresh_tokens` table. |
| MFA session token | 10 minutes | Bridges password login and `VerifyMFA`. |

**Refresh-token rotation.** `RefreshToken` runs in a database transaction: it marks the presented `jti` as revoked, inserts a new `jti`, and commits. A token that has already been rotated or revoked is rejected with `Unauthenticated`. `Logout` revokes the presented refresh token.

**MFA.**

- *TOTP:* `EnableMFA` with type `1` generates a 32-byte base32 secret, stores it in `mfa_secrets`, and returns the secret plus an `otpauth://` URL with issuer `goth`. Verification follows RFC 6238.
- *Email OTP:* `EnableMFA` with type `2`, and `SendOTP`, generate a six-digit code from `crypto/rand`, bcrypt-hash it, store the hash in Redis with a five-minute TTL, and email the plaintext code. `VerifyMFA` and `DisableMFA` compare the submitted code against the stored hash and delete the key on success.
- One MFA record per user. A second `EnableMFA` returns `AlreadyExists`.
- `VerifyMFA` writes the confirmed `mfa_type` to the `users` row and issues a fresh access/refresh token pair.

## Data storage

**PostgreSQL** (`migrations/001_init.sql`): `users`, `refresh_tokens`, `mfa_secrets`, `audit_logs`.

**Redis** keys, all with a five-minute TTL and a bcrypt hash as the value:

| Key | Set by |
|---|---|
| `mfa:email_otp:<user_id>` | `EnableMFA` (type 2), `SendOTP` (purpose MFA) |
| `password_reset:email_otp<user_id>` | `SendOTP` (purpose PASSWORD_RESET) |
| `acc_verify:email_otp<email>` | `SendOTP` (purpose EMAIL_VERIFY) |

## Repository layout

```
cmd/server/             Entrypoint: wiring for DB, Redis, interceptors, gRPC server
grpc/goth/              Generated Fory types, gRPC stubs, and CodecV2 (do not edit by hand)
schema/goth.fdl         Fory schema: messages, enums, and the GothService definition
migrations/             SQL schema
internal/db/            pgxpool wrapper
internal/redis/         go-redis client constructor
internal/handlers/      One file per RPC
internal/interceptors/  Recovery, logging, and bearer-token auth
internal/services/      JWT generation/validation, SMTP, TOTP URL
internal/helpers/       Env lookup, OTP and TOTP-secret generation, bcrypt helpers
client/cli/             gothtest CLI (separate Go module)
```
