FROM golang:1.19
WORKDIR /app

RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/ipoluianov/aneth_blocks_provider .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /aneth-block-provider
EXPOSE 8201
CMD ["/aneth-block-provider"]
