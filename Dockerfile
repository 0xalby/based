FROM golang:latest
WORKDIR /base
COPY . .
RUN go mod download
RUN go build -o bin/base -ldflags="-s -w" .
EXPOSE 16000
RUN useradd base
USER base
CMD ["./bin/base"]