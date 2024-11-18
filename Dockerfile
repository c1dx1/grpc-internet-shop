FROM golang:1.23 AS builder

WORKDIR /internet-shop

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o api-gateway ./services/api-gateway
RUN go build -o product-service ./services/product-service
RUN go build -o user-service ./services/user-service
RUN go build -o order-service ./services/order-service
RUN go build -o cart-service ./services/cart-service
RUN go build -o notification-service ./services/notification-service

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y libc6

WORKDIR /

COPY --from=builder /internet-shop/api-gateway /api-gateway
COPY --from=builder /internet-shop/product-service /product-service
COPY --from=builder /internet-shop/user-service /user-service
COPY --from=builder /internet-shop/order-service /order-service
COPY --from=builder /internet-shop/cart-service /cart-service
COPY --from=builder /internet-shop/notification-service /notification-service

EXPOSE 8080 50051 50052 50053 50054 50055
