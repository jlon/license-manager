# Stellar License Manager deployment

`runtime/` is intentionally ignored by Git. Create it on the server, copy `config.yaml.template` to `runtime/config.yaml`, and replace every `REPLACE_` value with a distinct random secret.

Create `runtime/mysql_password` and `runtime/mysql_root_password` with the same database password used in `runtime/config.yaml` and a separate MySQL root password. Generate the RSA keys into `runtime/` using the existing tool:

```bash
docker run --rm -v "$PWD:/src" -v "$PWD/deploy/stellar/runtime:/keys" \
  -w /src/backend golang:1.23 \
  go run ./cmd/gen-rsa-keys -private /keys/rsa_private_key.pem -public /keys/rsa_public_key.pem -size 4096
```

Build and start the isolated API and database:

```bash
docker compose -f deploy/stellar/docker-compose.yml up -d
curl --fail http://127.0.0.1:18888/health
```

Merge `nginx.conf.snippet` into the existing `www.xclaw.live` Nginx configuration, test it with `nginx -t`, then reload Nginx. The public API path is `https://www.xclaw.live/license/api/v1/stellar/trial`.
