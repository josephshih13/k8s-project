FROM golang:1.15.3-alpine
WORKDIR /backend
ADD . /backend
RUN cd /backend && go build -o backend
ENTRYPOINT ./backend