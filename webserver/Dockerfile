FROM golang:1.13.5 as builder

ARG arch

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN make build$arch

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/powerserver .
USER nonroot:nonroot