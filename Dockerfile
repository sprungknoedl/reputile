FROM alpine
LABEL maintainer "Thomas Kastner <tom@sprungknoedl.at>"

RUN apk add --no-cache ca-certificates

COPY reputile /app/reputile
COPY templates /app/templates
COPY static /app/static

ENV PORT=8080
EXPOSE 8080

WORKDIR /app
CMD ["/app/reputile"]
