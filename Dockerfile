FROM golang:latest
RUN apt-get update
RUN apt-get upgrade -y
RUN go get -u github.com/gofiber/fiber
RUN go get gorm.io/gorm
RUN go get github.com/satori/go.uuid
RUN go get gorm.io/driver/postgres

COPY ./main.go  .
RUN go build ./main.go

CMD ./main