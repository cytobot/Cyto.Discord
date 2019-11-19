FROM golang:1.13.4-alpine AS build-env

ENV GO111MODULE=on

WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build-env /app/ .

CMD [ "./main" ]