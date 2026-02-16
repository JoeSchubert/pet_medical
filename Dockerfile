# Stage 1: build frontend
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: build Go backend with embedded frontend
FROM golang:1.24-alpine AS backend
WORKDIR /app
RUN apk add --no-cache git
COPY backend/go.mod backend/go.sum ./
RUN go mod download 2>/dev/null || true
COPY backend/ ./
COPY --from=frontend /app/frontend/dist ./cmd/api/static/
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

# Stage 3: minimal runtime (Tesseract for document OCR/search)
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tesseract-ocr tesseract-ocr-data-eng
WORKDIR /app
COPY --from=backend /api .
EXPOSE 8080
ENV PORT=8080
CMD ["./api"]
