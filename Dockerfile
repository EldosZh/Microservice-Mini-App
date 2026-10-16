FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -buildvcs=false -o /app ./cmd/${SERVICE}

FROM alpine:3.21
RUN adduser -D -H app
USER app
COPY --from=build /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
