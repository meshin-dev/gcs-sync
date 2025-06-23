FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY . /src
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gcs-sync main.go

FROM google/cloud-sdk:alpine

WORKDIR /app
COPY --from=builder /bin/gcs-sync ./gcs-sync

ENTRYPOINT ["/bin/bash", "-c", "\
  gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS && \
  exec ./gcs-sync"]
