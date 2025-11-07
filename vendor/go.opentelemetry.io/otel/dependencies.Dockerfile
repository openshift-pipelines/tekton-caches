# This is a renovate-friendly source of Docker images.
FROM python:3.13.5-slim-bullseye@sha256:5b9fc0d8ef79cfb5f300e61cb516e0c668067bbf77646762c38c94107e230dbc AS python
FROM otel/weaver:v0.19.0@sha256:3d20814cef548f1d31f27f054fb4cd6a05125641a9f7cc29fc7eb234e8052cd9 AS weaver
FROM avtodev/markdown-lint:v1@sha256:6aeedc2f49138ce7a1cd0adffc1b1c0321b841dc2102408967d9301c031949ee AS markdown
