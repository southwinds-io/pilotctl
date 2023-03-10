#   Pilot Control Service
#   Copyright (C) 2022-Present SouthWinds Tech Ltd - www.southwinds.io
#
#   This program is free software: you can redistribute it and/or modify
#   it under the terms of the GNU Affero General Public License as published by
#   the Free Software Foundation, either version 3 of the License, or
#   (at your option) any later version.
#
#   This program is distributed in the hope that it will be useful,
#   but WITHOUT ANY WARRANTY; without even the implied warranty of
#   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#   GNU Affero General Public License for more details.
#
#   You should have received a copy of the GNU Affero General Public License
#   along with this program.  If not, see <https://www.gnu.org/licenses/>.
FROM registry.access.redhat.com/ubi8/ubi-minimal

LABEL author="SouthWinds Tech Ltd"
LABEL maintainer="admin@southwinds.io"
LABEL artisan.svc.manifest="/app/svc.yaml"

ARG UNAME=pilotctl

ENV UID=100
ENV GID=100

ENV PILOT_CTL_SYNC_PATH=/sync
ENV PILOT_CTL_TELEM_BUFFER_PATH=/telem-buffer

RUN microdnf install shadow-utils.x86_64 && \
    groupadd -g $GID -o $UNAME && \
    useradd -M -u $UID -g $GID $UNAME && \
    rm -rf /var/cache/yum && \
    microdnf clean all && \
    mkdir -p ${PILOT_CTL_SYNC_PATH} && chown -R ${UNAME} ${PILOT_CTL_SYNC_PATH} && \
    mkdir -p ${PILOT_CTL_TELEM_BUFFER_PATH} && chown -R ${UNAME} ${PILOT_CTL_TELEM_BUFFER_PATH}

USER $UNAME

WORKDIR /app

COPY pilotctl mongo-metrics mongo-logs pravega-metrics influx-metrics svc.yaml ./

# mount for PGP signing key
VOLUME /keys

# mount for event receivers configuration file
VOLUME /conf

# mount for sync files
VOLUME $PILOT_CTL_SYNC_PATH
# mount for telemetry connector buffer
VOLUME $PILOT_CTL_TELEM_BUFFER_PATH

CMD ["sh", "-c", "/app/pilotctl"]

EXPOSE 8080/tcp