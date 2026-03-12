FROM golang:1.26 AS build-stage
WORKDIR /app
COPY example-resource.go go.mod go.sum ./
COPY ini /usr/local/go/src/example-resource/ini
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /example-resource

FROM scratch AS release-stage
COPY --from=build-stage /example-resource /example-resource
COPY example-resource.ini.sample /example-resource.ini
ENTRYPOINT ["/example-resource"]
ENV PROVIDER_URL=
ENV LISTEN_ADDRESS=0.0.0.0:8080
