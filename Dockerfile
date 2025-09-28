FROM golang:1.25-alpine AS build

WORKDIR /app

COPY files /app/files/
COPY cmd /app/cmd
COPY internal /app/internal
COPY pkg /app/pkg

RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /app/main cmd/main/main.go


FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main /app/main
COPY --from=build /app/files/ /app/files/
COPY  static /app/static/
COPY templates /app/templates/

EXPOSE 8080

WORKDIR /app
CMD ["./main"]
