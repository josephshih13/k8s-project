FROM golang:1.15.3-alpine
WORKDIR /backend
ADD . /backend
RUN cd /backend && go build
ENTRYPOINT ./backend