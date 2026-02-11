# Running BillionMail Behind a Reverse Proxy

BillionMail can be placed behind a reverse proxy (nginx, Caddy, Traefik, etc.) to add SSL termination, custom domains, or integrate with existing infrastructure.

## The `reverse_proxy_domain` Setting

BillionMail stores a `reverse_proxy_domain` value in the `bm_options` table. When set, tracking URLs (open/click tracking) and public-facing links use this domain instead of the internal container address.

**Set via UI:** Settings > General > Reverse Proxy Domain

The domain must include the scheme, e.g. `https://mail.example.com`. BillionMail tests connectivity by calling `GET {domain}/api/languages/get` before saving.

## Required Headers

All proxy configs must forward these headers:

| Header | Purpose |
|--------|---------|
| `Host` | Original host header for correct URL generation |
| `X-Real-IP` | Client's real IP for analytics and rate limiting |
| `X-Forwarded-For` | Full proxy chain for logging |
| `X-Forwarded-Proto` | `https` — ensures BillionMail generates HTTPS tracking URLs |

## Nginx

```nginx
server {
    listen 443 ssl http2;
    server_name mail.example.com;

    ssl_certificate     /etc/ssl/certs/mail.example.com.pem;
    ssl_certificate_key /etc/ssl/private/mail.example.com.key;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:8080;

        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support (live updates)
        proxy_http_version 1.1;
        proxy_set_header Upgrade    $http_upgrade;
        proxy_set_header Connection "upgrade";

        # SPA fallback — let BillionMail handle routing
        proxy_intercept_errors off;
    }
}

# HTTP -> HTTPS redirect
server {
    listen 80;
    server_name mail.example.com;
    return 301 https://$host$request_uri;
}
```

## Caddy

```caddyfile
mail.example.com {
    reverse_proxy 127.0.0.1:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

Caddy handles SSL certificates automatically via Let's Encrypt.

## Traefik (Docker labels)

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.billionmail.rule=Host(`mail.example.com`)"
  - "traefik.http.routers.billionmail.tls.certresolver=letsencrypt"
  - "traefik.http.services.billionmail.loadbalancer.server.port=8080"
```

## After Setup

1. Set the reverse proxy domain in BillionMail: Settings > General > Reverse Proxy Domain
2. Enter your full URL with scheme: `https://mail.example.com`
3. BillionMail will test the connection before saving

## Troubleshooting

### SPA 404 errors on page refresh

The frontend is a Vue SPA — all routes must resolve to the BillionMail backend, not return static 404s. Ensure your proxy passes all paths to the upstream (no `try_files` that serve a local 404 page).

### Tracking URLs showing internal port (e.g. `:5679`)

The `reverse_proxy_domain` is not set, or `X-Forwarded-Proto` / `Host` headers are not forwarded. BillionMail falls back to the internal address. Set the reverse proxy domain in Settings after confirming headers are passed.

### SSL certificate mismatch

Ensure the certificate matches the domain used in `reverse_proxy_domain`. If using Caddy, certificates are automatic. For nginx, verify `ssl_certificate` matches `server_name`.

### WebSocket connection failures

If live dashboard updates stop working behind the proxy, ensure you forward `Upgrade` and `Connection` headers (see nginx config above). Caddy handles this automatically.
