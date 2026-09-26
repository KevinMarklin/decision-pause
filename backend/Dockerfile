# syntax=docker/dockerfile:1
# Контекст сборки — корень репозитория (backend/ + frontend/ рядом).

# --- frontend: Vite build ---
FROM node:24-alpine AS frontend
WORKDIR /web
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# --- backend: go build ---
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime ---
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=frontend /web/dist /app/dist
ENV FRONTEND_DIST=/app/dist
EXPOSE 8080
USER nobody
CMD ["/app/server"]
