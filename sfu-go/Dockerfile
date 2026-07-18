FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY main.go ./
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /sfu .

FROM alpine:3.20
COPY --from=build /sfu /usr/local/bin/sfu
ENV BABELGRID_ADDR=:8080
EXPOSE 8080
CMD ["sfu"]
