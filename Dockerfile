
FROM golang:1.15.0-alpine3.12 

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go get github.com/go-redis/redis/v8 
RUN go get -u github.com/gorilla/mux


RUN go build -o main .

CMD ["/app/main"]