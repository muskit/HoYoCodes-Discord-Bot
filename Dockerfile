# Build
FROM golang:alpine AS build

WORKDIR /build
COPY . .
RUN /build/scripts/build.sh

# Create img from binary
FROM alpine:latest as prod
LABEL org.opencontainers.image.source https://github.com/muskit/HoYoCodes-Discord-Bot
RUN apk add gcompat

WORKDIR /app
COPY --from=build /build/app .
CMD ["./app", "-debug"]
