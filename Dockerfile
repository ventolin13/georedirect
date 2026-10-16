FROM golang:1.25 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -buildvcs=false -o /out/georedirect ./cmd

FROM ubuntu:22.04

RUN apt update && apt install -y ca-certificates && update-ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/georedirect .
COPY deployments/prod.yaml .

ENV CONFIG=/app/prod.yaml

EXPOSE 8443/tcp

CMD ["./georedirect"]
