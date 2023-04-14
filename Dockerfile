FROM golang:1.20 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /app/chatemployee

FROM alpine:latest

RUN apk add --no-cache tzdata
ENV TZ=Asia/Ho_Chi_Minh

WORKDIR /app
COPY --from=build /app/chatemployee /app/chatemployee
COPY ./config.yaml /app/config.yaml

EXPOSE 9000

ENTRYPOINT ["/app/chatemployee"]
