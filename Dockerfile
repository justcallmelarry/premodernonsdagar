FROM golang:alpine AS build

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /app/main cmd/main/main.go


FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main /app/main
COPY --from=build /app/files/ /app/files/
COPY --from=build /app/static/ /app/static/
COPY --from=build /app/templates/ /app/templates/

EXPOSE 8080

WORKDIR /app
CMD ["./main"]
