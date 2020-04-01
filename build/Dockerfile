FROM alpine:3.9.6

ENV APP_DIR=/sts-annotator \
    APP_USER="sts-annotator" \
    APP_GROUP="sts-annotator"

ADD ./sts-annotator ${APP_DIR}/

USER root

# Create a group and user to run our app
RUN \
    adduser -D $APP_USER && addgroup $APP_USER $APP_GROUP && \
    chown -R ${APP_USER}:root $APP_DIR && \
    chmod -R ug+rwx $APP_DIR

WORKDIR $APP_DIR

USER $APP_USER

CMD ["./sts-annotator"]

