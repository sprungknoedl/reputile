FROM alpine
LABEL maintainer "Thomas Kastner <tom@sprungknoedl.at>"

ENV PORT=8080
EXPOSE 8080

RUN apk add --no-cache ca-certificates
RUN adduser -D -h /app reputile
USER reputile

WORKDIR /app
CMD ["/app/reputile"]

COPY reputile /app/reputile
COPY templates /app/templates
COPY static /app/static
