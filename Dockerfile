FROM alpine:latest

# system dependencies
RUN apk add gcompat

WORKDIR /app
COPY app /app/app
ENTRYPOINT [ "/app/app" ]