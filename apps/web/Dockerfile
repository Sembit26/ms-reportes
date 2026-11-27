# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM node:22.22.0-alpine3.23@sha256:e4bf2a82ad0a4037d28035ae71529873c069b13eb0455466ae0bc13363826e34 AS web-build
WORKDIR /workspace
RUN corepack enable && corepack prepare pnpm@10.13.1 --activate
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml .npmrc ./
COPY apps/web/package.json apps/web/
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --frozen-lockfile --ignore-scripts
COPY apps/web/ apps/web/
RUN pnpm --filter @pingeso-civil/ms-reportes-ui build

FROM --platform=$BUILDPLATFORM golang:1.26.11-alpine3.23@sha256:23f198931d4bd83b4f5a613bd75656257ae8bf832dcbc2db80e1050b5219ca2a AS server-build
WORKDIR /src
COPY apps/web/spa-server/ ./
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/spa-server .

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6
WORKDIR /srv
COPY --from=web-build --chown=nonroot:nonroot /workspace/apps/web/dist/ /srv/
COPY --from=server-build --chown=nonroot:nonroot /out/spa-server /app/spa-server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/spa-server"]
