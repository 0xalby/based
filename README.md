# Base
> Monoid API handling authentication with email verification and 2FA TOTP

## Motivation
Enables me to just think about my actual server logic to solve whatever problem I am trying to solve and nothing else.

I like Tiago's idea of moving to a micro service infrastracure(gRPC, proto buffers and a message borkers like RabbitMQ and Kafka) after whatever you are building is successful.

## Installation
You can get the executable from the [release page](https://github.com/0xalby/base/releases) ### Using Go's package manager
```
go install github.com/0xalby/base@latest
```
### From source
```
git clone https://github.com/0xalby/base
cd base
go mod tidy
make build
```
### Using Docker
```
make docker
```

## Usage
* Not setting SMTP_ADDRESS will skip account verification via email