FROM alpine:latest
RUN apk --no-cache add iproute2 iptables
RUN mkdir -p /opt/shaper
WORKDIR /opt/shaper
ENTRYPOINT [ "/sbin/shaper" ]