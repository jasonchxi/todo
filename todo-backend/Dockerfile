FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM golang:1.17
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8000
CMD ["./main"]
