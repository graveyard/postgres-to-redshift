FROM golang:1.5
RUN apt-get update
RUN apt-get install -y curl build-essential

# Gearcmd
RUN curl -L https://github.com/Clever/gearcmd/releases/download/v0.3.3/gearcmd-v0.3.3-linux-amd64.tar.gz | tar xz -C /usr/local/bin --strip-components 1

# copy source code
RUN mkdir -p /go/src/github.com/Clever/postgres-to-redshift
ADD . /go/src/github.com/Clever/postgres-to-redshift

WORKDIR /go/src/github.com/Clever/postgres-to-redshift

RUN go install github.com/Clever/postgres-to-redshift
RUN go build -o /usr/local/bin/postgres-to-redshift github.com/Clever/postgres-to-redshift

CMD ["/go/src/github.com/Clever/postgres-to-redshift/run_as_worker.sh"]
