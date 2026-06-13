# --- Stage 1: build the SvelteKit frontend (static SPA) ---
FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- Stage 2: build the Go server (static binary, migrations embedded) ---
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/nutcracker ./cmd/nutcracker

# --- Stage 3: minimal runtime image ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /bin/nutcracker /app/nutcracker
COPY --from=frontend /web/build /app/web/build
ENV STATIC_DIR=/app/web/build
EXPOSE 8080
CMD ["/app/nutcracker"]
