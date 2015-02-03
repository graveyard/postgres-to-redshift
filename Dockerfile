FROM golang:1.4
RUN apt-get update
RUN apt-get install -y curl build-essential
RUN go get github.com/tools/godep

# Gearcmd
RUN curl -L https://github.com/Clever/gearcmd/releases/download/v0.3.1/gearcmd-v0.3.1-linux-amd64.tar.gz | tar xz -C /usr/local/bin --strip-components 1

# copy source code
RUN mkdir -p /go/src/github.com/Clever/postgres-to-redshift
ADD . /go/src/github.com/Clever/postgres-to-redshift

# set workdir to find saved godeps
WORKDIR /go/src/github.com/Clever/postgres-to-redshift

# build source code using godep
RUN rm -rf /go/src/github.com/Clever/postgres-to-redshift/Godeps/_workspace/pkg/
RUN godep go install github.com/Clever/postgres-to-redshift
RUN godep go build -o /usr/local/bin/postgres-to-redshift github.com/Clever/postgres-to-redshift

CMD ["/go/src/github.com/Clever/postgres-to-redshift/run_as_worker.sh"]
