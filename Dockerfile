FROM golang:1.14-alpine as builder
ENV GO111MODULE=on
LABEL maintainer="Johann Wagner <johann@wagnerdevelopment.de>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch
COPY --from=builder /app/certmanager-interoperability-operator /usr/local/bin/certmanager-interoperability-operator
ENTRYPOINT ["/usr/local/bin/certmanager-interoperability-operator"]