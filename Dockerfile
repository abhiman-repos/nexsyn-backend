# Build Stage
FROM golang:1.26-alpine AS builder


WORKDIR /nexcyn-backend

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN go build -o server ./cmd/server

#final stage
FROM alpine:latest

WORKDIR /nexcyn-backend

COPY --from=builder /app/server .

EXPOSE 3000

CMD [ "./server" ]
