FROM alpine:3.12

LABEL maintainer="jenkins-x"

# Slack connector, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# OpenSSL is required so wget can query HTTPS endpoints for health checking.
RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates curl git netcat-openbsd openssl

RUN git config --global credential.helper store

EXPOSE 8080
ENTRYPOINT ["/jx-slack"]
CMD ["run"]

COPY ./build/linux/ /