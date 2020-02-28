FROM golang:1.13.5 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN make build

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/sensor .
