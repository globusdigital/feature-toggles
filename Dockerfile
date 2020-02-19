FROM golang:alpine3.10 AS builder
MAINTAINER vkojouharov

RUN apk add --no-cache ca-certificates

COPY . /feature-toggles
WORKDIR /feature-toggles

RUN go mod download

RUN go build -ldflags "-s -w" ./cmd/feature-toggles

FROM golang:alpine3.10 AS runtime
MAINTAINER vkojouharov

COPY --from=builder /feature-toggles/feature-toggles /usr/local/bin

EXPOSE 80

CMD ["feature-toggles"]
