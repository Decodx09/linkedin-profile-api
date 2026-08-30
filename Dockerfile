FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 appuser
USER appuser
WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8000
ENV PORT=8000
ENTRYPOINT ["/app/server"]
