FROM golang:1.23.0

RUN mkdir /library
WORKDIR /library

COPY . .
RUN chmod a+x docker/*.sh