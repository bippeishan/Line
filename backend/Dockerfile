FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /line-api ./cmd/api

FROM alpine:3.20
COPY --from=build /line-api /line-api
EXPOSE 8080
CMD ["/line-api"]
