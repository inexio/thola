FROM golang:latest AS build
WORKDIR /go/src/thola
COPY . .
RUN CGO_ENABLED=0 go build -v -o thola .

FROM alpine:latest
RUN apk add --no-cache bash
COPY --from=build /go/src/thola/thola .

ENTRYPOINT ["./thola", "api"]
