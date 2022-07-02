FROM golang:1.18-alpine as build
RUN apk upgrade --no-cache --force
RUN apk add --update build-base make git
WORKDIR /go/src/github.com/vibin18/whatsapp_crawler

# Compile
COPY ./ /go/src/github.com/vibin18/whatsapp_crawler
RUN make dependencies
RUN make build
RUN ./whatsapp_crawler --help

# Final Image
FROM gcr.io/distroless/static AS export-stage
ADD user.yaml /
COPY --from=build /go/src/github.com/vibin18/whatsapp_crawler/whatsapp_crawler /