FROM golang:latest

WORKDIR /
COPY . ./

RUN go build -o misiek

CMD [ "./misiek" ]