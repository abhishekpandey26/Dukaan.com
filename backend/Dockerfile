# ---- Stage 1: build the Go binary ----
FROM golang:1.27-alpine AS build
WORKDIR /app
# Copy go.mod/go.sum first so Docker can cache downloaded modules.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 -> a fully static binary that runs on a tiny base image.
RUN CGO_ENABLED=0 go build -o /gocommerce .

# ---- Stage 2: tiny runtime image (~20 MB instead of ~800 MB) ----
FROM alpine:3.22
RUN adduser -D app
USER app
COPY --from=build /gocommerce /gocommerce
EXPOSE 8080
CMD ["/gocommerce"]
