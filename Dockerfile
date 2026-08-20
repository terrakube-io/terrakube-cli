FROM dhi.io/golang:1.26-debian13-dev AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X terrakube/cmd.Version=${VERSION} -X terrakube/cmd.Commit=${COMMIT} -X terrakube/cmd.Date=${DATE}" -o /terrakube .

FROM dhi.io/alpine-base:3.23
# hadolint ignore=DL3018
# RUN apk add --no-cache ca-certificates
COPY --from=build /terrakube /usr/local/bin/terrakube
ENTRYPOINT ["terrakube"]
