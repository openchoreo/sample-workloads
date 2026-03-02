#!/bin/sh
set -eu

: "${DOC_SERVICE_URL:?missing DOC_SERVICE_URL}"
: "${COLLAB_SERVICE_URL:?missing COLLAB_SERVICE_URL}"

envsubst '$DOC_SERVICE_URL $COLLAB_SERVICE_URL' < /etc/nginx/nginx.conf.template > /etc/nginx/conf.d/default.conf
exec nginx -g 'daemon off;'
