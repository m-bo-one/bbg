FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
  build-essential \
  libmysqlclient-dev \
  python3-pip \
  python3-dev \
  nodejs-legacy \
  npm \
  wget \
  autoconf \
  autogen \
  golang-goprotobuf-dev \
  autoconf \
  automake \
  libtool \
  curl \
  make \
  g++ \
  unzip \
  pkg-config \
  libgtk-3-dev

RUN cd /usr/local/bin \
  && ln -s /usr/bin/python3 python \
  && pip3 install --upgrade pip \
  && npm install -g gulp yarn

# # protoc3
RUN wget "https://github.com/google/protobuf/releases/download/v3.4.1/protobuf-cpp-3.4.1.tar.gz" -O /tmp/protobuf.tar.gz \
  && tar -xvzf /tmp/protobuf.tar.gz -C /tmp \
  && cd /tmp/protobuf-3.4.1 \
  && ./autogen.sh \
  && ./configure --prefix=/usr/local \
  && make -j 4 \
  && make install \
  && ldconfig

RUN curl -sL https://deb.nodesource.com/setup_6.x | bash - \
  && apt-get install -y --no-install-recommends nodejs \
  && npm install -g gulp yarn

WORKDIR /client