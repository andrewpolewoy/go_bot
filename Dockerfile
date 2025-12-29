FROM --platform=$BUILDPLATFORM golang:1.24 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# BuildKit provides TARGETOS/TARGETARCH for multi-platform builds.
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/bot ./cmd/bot

FROM gcr.io/distroless/base-debian12
WORKDIR /

COPY --from=build /out/bot /bot

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/bot"]
