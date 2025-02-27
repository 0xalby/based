# Based
> Monoid API handling authentication with email verification and 2FA TOTP

## Motivation
Enables me to focus on my actual server side application logic, design and infrastructure as the API takes care of the rest.

The "rest" being problems already solved such as authentication and database CRUD operations.

I like Tiago's idea of moving to a micro service infrastracure(gRPC, proto buffers and a message borkers like RabbitMQ and Kafka) after whatever you are building is successful that is why this is a monoid which also comes in handy if you just wanna try an idea out.

## Features
* SQLite3 and Postgres support(more to come in the future)
* Authentication(JWT, 2FA TOTP and optional email verification)
* Single static executable
* Modular with dependency injections
* Commented all the way
* Audited for BOLA, CSRF, XSS and SQL injections

## Installation
1. You can get an executable from the [release page](https://github.com/0xalby/based/releases)
2. Using Go's package manager
```zsh
go install github.com/0xalby/based@latest
```
3. From source
```zsh
git clone https://github.com/0xalby/based
cd based
go mod tidy
vim .env
make up
make build
make run
```
4. In a Docker container
```zsh
vim .env
make docker
docker run --env-file .env -p 8080:16000 --volume log:log based
```

## Utilities
```zsh
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Reference
### Auth
```zsh
# Register
curl -X POST http://localhost:16000/register \
-H "Content-Type: application/json" \
-d '{
  "email": "user@example.com",
  "password": "securepassword123"
}'

# Login
curl -X POST http://localhost:16000/login \
-H "Content-Type: application/json" \
-d '{
  "email": "user@example.com",
  "password": "securepassword123",
  "totp": "123456" # Optional, only if TOTP is enabled
}'

# Email verification
curl -X POST http://localhost:16000/verification \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <JWT_TOKEN>" \
-d '{
  "code": "123456"
}'

# Resend email verification
curl -X POST http://localhost:16000/resend \
-H "Authorization: Bearer <JWT_TOKEN>" \

# Login with a 2FA(TOTP) backup code
curl -X POST http://localhost:16000/api/v1/auth/backup \
-H "Content-Type: application/json" \
-d '{
  "email": "user@example.com",
  "backup_code": "ABCD-EFGH"
}'# 
```
### Account
```zsh
# Send account changes confirmation email
curl -X GET http://localhost:16000/api/v1/account/confirmation \
-H "Authorization: Bearer <JWT_TOKEN>"

# Update account email
curl -X PUT http://localhost:16000/api/v1/account/update/email \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <JWT_TOKEN>" \
-d '{
  "email": "newuser@example.com"
}'

# Update account password
curl -X PUT http://localhost:16000/api/v1/account/update/password \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <JWT_TOKEN>" \
-d '{
  "password": "newsecurepassword123"
}'

# Enabling 2FA(TOTP)
curl -X PUT http://localhost:16000/api/v1/account/totp/enable \
-H "Authorization: Bearer <JWT_TOKEN>"

# Disabling 2FA(TOTP)
curl -X PUT http://localhost:16000/api/v1/account/totp/disable \
-H "Authorization: Bearer <JWT_TOKEN>"

# Deleting account
curl -X DELETE http://localhost:16000/api/v1/account/delete \
-H "Authorization: Bearer <JWT_TOKEN>"

# Recovery
curl -X GET http://localhost:16000/api/v1/account/recovery \
-H "Content-Type: application/json" \
-d '{
  "email": "user@example.com"
}'

# Password reset
curl -X POST http://localhost:16000/api/v1/account/reset \
-H "Content-Type: application/json" \
-d '{
  "email": "user@example.com",
  "code": "123456",
  "password": "newsecurepassword123"
}'
```

## Contributing
Check out [TODO.md](./TODO.md) and send a PR for me to review