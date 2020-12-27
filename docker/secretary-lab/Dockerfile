FROM golang:1.15 AS builder
ENV PROJECT_DIR /go/src/github.com/chez-shanpu/secretary
COPY ./ ${PROJECT_DIR}/
WORKDIR ${PROJECT_DIR}/cmd/secretary-lab
RUN go build -i -o ${PROJECT_DIR}/bin/secretary-lab


FROM ubuntu:focal AS prod
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/src/github.com/chez-shanpu/secretary/bin/secretary-lab /
CMD ["/secretary-lab"]