FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY .. ./

RUN go build -o main ./cmd/tic_tac_toe/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/db/cred_db.yaml ./db/cred_db.yaml

CMD ["./main"]
