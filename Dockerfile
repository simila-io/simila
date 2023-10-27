FROM golang:1.19.1 as builder

WORKDIR /usr/src/simila

COPY . .

RUN apt update && apt -y --no-install-recommends install openssh-client git && \
        mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts && \
        git config --global url."git@github.com:".insteadOf "https://github.com"

RUN go env -w GOPRIVATE="github.com/simila-io/*"

RUN --mount=type=ssh CGO_ENABLED=0 make all

FROM alpine:3.16

ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.12/grpc_health_probe-linux-amd64 /bin/grpc_health_probe

RUN chmod +x /bin/grpc_health_probe

EXPOSE 50051

WORKDIR /app

COPY --from=builder /usr/src/simila/build/simila .
COPY --from=builder /usr/src/simila/config/simila.yaml .

CMD ["/app/simila", "start", "--config", "/app/simila.yaml"]
