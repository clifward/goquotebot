# syntax=docker/dockerfile:1

# ------------------------------------------------------
# Production Build
# ------------------------------------------------------

FROM golang:1.17.6-buster as build

WORKDIR /home/pi/Goquotebot/goquotebot

RUN apt-get update -y && apt-get install build-essential -y

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN cd cmd/goquote && CGO_ENABLED=1 GOARM=5 GOOS=linux GOARCH=arm go build -o /goquote -ldflags="-w -s" .

CMD [ "/goquote" ]

# ------------------------------------------------------
# Production Deploy
# ------------------------------------------------------


FROM golang:1.17.6-buster

WORKDIR /home/pi/Goquotebot/goquotebot

COPY --from=build /goquote /goquote
COPY pkg/telegram/templates/ pkg/telegram/templates/
COPY config.yaml .

EXPOSE 8080

ENTRYPOINT ["/goquote"]
