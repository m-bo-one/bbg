FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
  build-essential \
  libmysqlclient-dev \
  python3-pip \
  python3-dev \
  nodejs-legacy \
  npm \
  wget \
  golang-goprotobuf-dev

RUN cd /usr/local/bin \
  && ln -s /usr/bin/python3 python \
  && pip3 install --upgrade pip \
  && npm install -g gulp yarn

WORKDIR /client