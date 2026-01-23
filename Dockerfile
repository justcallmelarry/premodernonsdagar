FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY cmd /app/cmd
COPY internal /app/internal
COPY pkg /app/pkg
COPY files /app/files
COPY input /app/input/

RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /app/main cmd/main/main.go
RUN /app/main --build


FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main /app/main
COPY --from=build /app/input /app/input
COPY --from=build /app/files /app/files
COPY static /app/static/
COPY templates /app/templates/

EXPOSE 8080

WORKDIR /app
CMD ["./main"]
