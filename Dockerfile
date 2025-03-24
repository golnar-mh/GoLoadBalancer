FROM golang:1.20
WORKDIR /app
COPY . .
RUN go mod init loadbalancer
RUN go get github.com/gorilla/mux
RUN go build -o loadbalancer
CMD ["./loadbalancer"]