#!/bin/sh
# Generates /etc/nginx/conf.d/default.conf from the template by substituting
# the __SUBPATH__ placeholder with $NGINX_BASE_PATH, then starts nginx.
#
# Examples:
#   NGINX_BASE_PATH=""            -> served at /          (default, root)
#   NGINX_BASE_PATH="/dragncards" -> served at /dragncards/
set -e

BASE="${NGINX_BASE_PATH:-}"

# Normalise: strip any trailing slash (the template already appends one).
BASE="${BASE%/}"

# Ensure a leading slash when non-empty so nginx is happy.
if [ -n "$BASE" ] && [ "${BASE#/}" = "$BASE" ]; then
  BASE="/$BASE"
fi

HREF="${BASE}/"
echo "nginx: serving frontend at '${HREF}'"

# Generate nginx config
sed "s|__SUBPATH__|${BASE}|g" \
    /etc/nginx/nginx.conf.template \
    > /etc/nginx/conf.d/default.conf

# Inject <base href> into index.html so React Router and relative assets
# resolve correctly at runtime without a rebuild
sed -i "s|__BASE_HREF__|${HREF}|g" \
    /usr/share/nginx/html/index.html

exec nginx -g "daemon off;"
