FROM golang:latest

WORKDIR /go/src/thola
COPY . .

RUN go get -d -v .
RUN go generate
RUN go install -v .

ENTRYPOINT ["thola", "api"]