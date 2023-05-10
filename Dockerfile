FROM golang:1.19-alpine
WORKDIR /usr/src/app

RUN apk add make && apk add build-base && apk add vips-dev

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# RUN make build

EXPOSE 8080

CMD ["make", "run"]