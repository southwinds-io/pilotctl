#
#    MongoDb Event Receiver for Pilot Control
#    © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
#    Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
#    Contributors to this project, hereby assign copyright in this code to the project,
#    to be licensed under the same terms as the rest of the code.
#
FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL author="skipper"
LABEL maintainer="skipper@southwinds.io"

ARG UNAME=mongoevr

ENV UID=100000000
ENV GID=100000000

RUN microdnf update --disablerepo=* --enablerepo=ubi-8-appstream --enablerepo=ubi-8-baseos -y && \
    microdnf install shadow-utils.x86_64 && \
    groupadd -g $GID -o $UNAME && \
    useradd -M -u $UID -g $GID $UNAME && \
    rm -rf /var/cache/yum && \
    microdnf clean all

USER $UNAME

WORKDIR /app

COPY ./bin/mongo-evr ./

CMD ["sh", "-c", "/app/mongo-evr"]

EXPOSE 8080/tcp