# syntax=docker/dockerfile:1

# The build context is this directory — see `build.context` in
# docker-compose.yml.

FROM golang:1.25-alpine AS builder

WORKDIR /src

# Dependencies first so a code-only change does not re-download the module
# cache.
#
# go.sum* with the glob, matching what the frontend Dockerfile already does for
# package-lock.json. A module with no external dependencies has no go.sum, and
# that is a correct state, not a missing file — run_go_mod_tidy says so in as
# many words when it produces none. Without the glob, COPY fails outright and
# the error names a line in this file rather than anything about the project.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# ./cmd/api, not ./... — and the difference is not style.
#
# `go build -o <file>` accepts exactly ONE package. Not one main package: one
# package. This line used to say ./..., on the reasoning that a project has a
# single entry point anyway, but ./... matches every package in the module, so
# the first backend to grow an internal/ package broke the image build with
#
#   go: cannot write multiple packages to non-directory /bin/server
#
# and every Go backend worth deploying has one. Measured on a generated project
# whose module held cmd/api, internal/db and migrations: three packages, and the
# build failed on a module where nothing was wrong.
#
# cmd/api is the entry point the scaffold instructions mandate, so naming it
# here is not a new constraint — it is the existing one, enforced where it can
# fail loudly instead of silently matching too much.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/api

# ── Runtime ──────────────────────────────────────────────────────────────────

FROM alpine:3.20

# ca-certificates for outbound TLS, tzdata so timestamps are not all UTC by
# accident, wget for the healthcheck in docker-compose.yml.
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /bin/server /app/server

ENV PORT=8080

USER nobody

EXPOSE 8080

CMD ["/app/server"]
