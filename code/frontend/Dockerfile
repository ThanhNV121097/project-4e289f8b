# syntax=docker/dockerfile:1

# The build context is this directory, not the repository root — see the
# `build.context` in docker-compose.yml. Every path below is relative to
# code/frontend/, which is the whole reason this file is fixed rather than
# rewritten per project: the version that got the context wrong looked
# completely reasonable and never built.

FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./

# --include=dev is load-bearing. npm omits devDependencies whenever NODE_ENV is
# production, and a Next.js app keeps typescript, tailwind and its @types
# there. Without typescript installed, `next build` cannot read tsconfig, so it
# silently drops the "@/*" path aliases and fails every import that uses one —
# "Module not found: Can't resolve '@/components/Header'", for a file sitting
# right there on disk.
RUN npm ci --include=dev

COPY . .

# public/ is optional in a Next.js app but the runtime stage copies it
# unconditionally, and COPY fails on a path that does not exist.
RUN mkdir -p public

ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# ── Runtime ──────────────────────────────────────────────────────────────────
# Requires `output: "standalone"` in next.config. That emits a self-contained
# server.js with only the modules it actually uses, which is what runs below.
# `next start` also works but prints a warning telling you not to, and it drags
# the full node_modules along.

FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

COPY --from=builder --chown=node:node /app/public ./public
COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static

USER node

EXPOSE 3000

CMD ["node", "server.js"]
