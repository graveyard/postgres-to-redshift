FROM debian:jessie
RUN apt-get update && \
    apt-get install -y curl build-essential

RUN curl -L https://github.com/Clever/gearcmd/releases/download/0.10.0/gearcmd-v0.10.0-linux-amd64.tar.gz | tar xz -C /usr/local/bin --strip-components 1

COPY bin/postgres-to-redshift /usr/local/bin/postgres-to-redshift
CMD ["gearcmd", "--name", "postgres-to-redshift", "--cmd", "/usr/local/bin/postgres-to-redshift"]
