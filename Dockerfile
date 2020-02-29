FROM alpine:3.8
# Slack connector, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# OpenSSL is required so wget can query HTTPS endpoints for health checking.
RUN apk add --update ca-certificates openssl

EXPOSE 8080
ENTRYPOINT ["/app-slack"]
COPY ./bin/ /