FROM alpine
LABEL maintainer "Thomas Kastner <tom@sprungknoedl.at>"

RUN apk add --no-cache ca-certificates

COPY reputile /app/reputile
COPY static /app/static
COPY templates /app/templates

ENV PORT=8080
EXPOSE 8080

WORKDIR /app
ENTRYPOINT ["/app/reputile"]
