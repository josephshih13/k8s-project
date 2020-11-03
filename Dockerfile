FROM golang:1.15.3-alpine
WORKDIR /backend
ADD . /backend
RUN cd /backend && go build -o backend
EXPOSE 9936
ENTRYPOINT ./backend