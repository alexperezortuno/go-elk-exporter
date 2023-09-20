FROM golang:1.19

RUN mkdir -p /root/elastic-query-export
COPY . /root/elastic-query-export

RUN cd /root/elastic-query-export \
  && go mod tidy \
  && go build
