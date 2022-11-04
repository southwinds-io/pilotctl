#
#    Pilot Control Service
#    © 2018-Present - SoutWinds Tech Ltd - www.southwinds.io
#    Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
#    Contributors to this project, hereby assign copyright in this code to the project,
#    to be licensed under the same terms as the rest of the code.
#
FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL author="skipper"
LABEL maintainer="skipper@southwinds.io"
LABEL artisan.svc.manifest="/app/svc.yaml"

ARG UNAME=pilotctl

ENV UID=100
ENV GID=100

ENV SYNC_PATH=/sync

RUN microdnf install shadow-utils.x86_64 && \
    groupadd -g $GID -o $UNAME && \
    useradd -M -u $UID -g $GID $UNAME && \
    rm -rf /var/cache/yum && \
    microdnf clean all && \
    mkdir -p ${SYNC_PATH} && \
    chown -R ${UNAME} ${SYNC_PATH}

USER $UNAME

WORKDIR /app

COPY pilotctl mongo-metrics pravega-metrics svc.yaml ./

# mount for PGP signing key
VOLUME /keys

# mount for event receivers configuration file
VOLUME /conf

# mount for sync files
VOLUME /sync

CMD ["sh", "-c", "/app/pilotctl"]

EXPOSE 8080/tcp