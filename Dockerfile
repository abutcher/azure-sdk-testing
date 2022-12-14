ARG BUILDER=mcr.microsoft.com/oss/go/microsoft/golang:1.19-bullseye
ARG BASEIMAGE=gcr.io/distroless/static:nonroot

FROM ${BUILDER} as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

# Build
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o azuresdktesting .

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM --platform=linux/amd64 ${BASEIMAGE}
WORKDIR /
COPY --from=builder /workspace/azuresdktesting .
# Kubernetes runAsNonRoot requires USER to be numeric
USER 65532:65532

ENTRYPOINT ["/azuresdktesting"]
