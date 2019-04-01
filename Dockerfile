FROM golang:1.12-stretch

WORKDIR /tmp
ENV DEP_VERSION v0.5.1
COPY dep-linux-amd64.sha256 .
RUN curl -sSLO https://github.com/golang/dep/releases/download/${DEP_VERSION}/dep-linux-amd64 && \
    sha256sum -c dep-linux-amd64.sha256 && \
    mv dep-linux-amd64 /usr/local/bin/dep && \
    chmod 755 /usr/local/bin/dep

RUN apt-get update && \
    apt-get install -y libsystemd-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR $GOPATH/src/github.com/sky-uk/owl
COPY . .
ENTRYPOINT ["make", "clean"]
CMD ["test"]
