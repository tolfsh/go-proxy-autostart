FROM golang:1.19 AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /app ./...

FROM busybox
COPY --from=build /app .

CMD ["/app"]
