#!/bin/sh

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &

sleep 2

/app/caddy-proxy-manager

