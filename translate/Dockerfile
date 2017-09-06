FROM octoblu/alpine-ca-certificates
MAINTAINER laputa <justlaputa@gmail.com>

COPY translate /translate
COPY slack-status /slack-status
COPY entry.sh /entry.sh

ENTRYPOINT "/entry.sh"
