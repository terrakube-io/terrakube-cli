FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /terrakube .

FROM alpine:3
# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates
COPY --from=build /terrakube /usr/local/bin/terrakube
ENTRYPOINT ["terrakube"]
