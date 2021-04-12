# use scratch for K8S
ARG BASEIMAGE=scratch

FROM golang:1.15 as builder
WORKDIR /app
COPY . /app
RUN STATIC=1 make build

# NOTE: CF requires more than scratch
# while K8S is fine with it.
# To build image for CF:
# docker build -t <iamge name> --build-arg BASEIMAGE=alpine:3.11 .

FROM $BASEIMAGE
COPY --from=builder /app/bin/server /
#RUN groupadd -r t2m && t2m --no-log-init -r -g t2m t2m
#USER nobody
EXPOSE 8080
CMD [ "/server" ]