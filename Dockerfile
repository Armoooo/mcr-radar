FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /mcr-radar ./cmd/mcr-radar

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /mcr-radar /mcr-radar
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/mcr-radar"]
