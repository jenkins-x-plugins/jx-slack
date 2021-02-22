FROM gcr.io/jenkinsxio/jx-cli-base:0.0.38

# Slack connector, such as GitHub and Google logins require root certificates.
# Proper installations should manage those certificates, but it's a bad user
# experience when this doesn't work out of the box.
#
# OpenSSL is required so wget can query HTTPS endpoints for health checking.
#RUN apk add --update ca-certificates openssl

RUN git config --global credential.helper store

EXPOSE 8080
ENTRYPOINT ["/jx-slack"]
CMD ["run"]

COPY ./build/linux/ /