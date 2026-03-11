# -------------------------------------------------------
# Stage 1: Build Vue frontend
# -------------------------------------------------------
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# -------------------------------------------------------
# Stage 2: Build Go backend (with embedded frontend)
# -------------------------------------------------------
FROM golang:1.23-alpine AS go-builder
WORKDIR /app

COPY backend/ ./
# Copy built frontend into the embed path
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN go mod download
# Build without -tags dev so embed_prod.go is compiled
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /svcompare .

# -------------------------------------------------------
# Stage 3: Minimal runtime (scratch + certs)
# -------------------------------------------------------
FROM scratch
COPY --from=go-builder /svcompare /svcompare
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENV DATABASE_PATH=/data/svcompare.db
ENV PORT=8080
ENV GO_ENV=production

ENTRYPOINT ["/svcompare"]
