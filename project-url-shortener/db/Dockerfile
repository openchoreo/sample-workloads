# Custom postgres image with schema baked in.
# This avoids volume-mount complexity for demo purposes
FROM postgres:15-alpine
COPY init.sql /docker-entrypoint-initdb.d/001_init.sql
