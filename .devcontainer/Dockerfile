FROM mcr.microsoft.com/devcontainers/base:bullseye

RUN apt-get update \
    && apt-get -y --no-install-recommends install ca-certificates gnupg netcat curl make\
    && apt-get update \
    && apt-get autoremove \
    && echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | tee /etc/apt/sources.list.d/goreleaser.list \
    && apt-get update \
    && apt-get install -y goreleaser