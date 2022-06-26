FROM golang:latest AS compiling_stage
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
COPY ./ ./
RUN go install .

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Alexander Nurullov<web_soft@mail.ru>"
WORKDIR /root/
COPY --from=compiling_stage /go/bin/pipeline .
ENTRYPOINT ./pipeline