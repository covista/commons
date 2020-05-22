FROM postgres:11

# Silence debconf TERM messages
RUN echo "debconf debconf/frontend select Noninteractive" | debconf-set-selections
RUN apt-get update && apt-get install -y \
      postgresql-11-postgis-3 \
      postgresql-plpython3-11 \
      python3-dev \
      python3-pip \
      # Utilities
      software-properties-common \
      apt-transport-https \
      ca-certificates \
      gnupg \
      wget

# Setup postgresql data dir and config
RUN mkdir /data/
RUN cat /usr/share/postgresql/postgresql.conf.sample > /data/postgresql.conf

# Add timescale
RUN sh -c "echo 'deb https://packagecloud.io/timescale/timescaledb/debian/ `lsb_release -c -s` main' > /etc/apt/sources.list.d/timescaledb.list"
RUN wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | apt-key add -
RUN apt-get update && apt-get install -y timescaledb-postgresql-11
RUN timescaledb-tune --quiet --yes --conf-path=/data/postgresql.conf

RUN python3 -m pip install openlocationcode

# Add initialization scripts
ADD setup.sql /docker-entrypoint-initdb.d/

RUN pg_lsclusters

# PostgreSQL ENV variables
ENV POSTGRES_USER=covid19
ENV POSTGRES_PASSWORD=covid19databasepassword
ENV POSTGRES_HOST_AUTH_METHOD=password
ENV POSTGRES_DB=covid19
ENV PGDATA=/data/

VOLUME /data
EXPOSE 5432
