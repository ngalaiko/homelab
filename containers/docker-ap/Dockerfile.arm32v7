FROM apcheamitru/arm32v7-alpine:3.7

RUN apk add --update --no-cache \
        bash \
        hostapd \
        iptables \
        dhcp \
        docker \
        iproute2 \
        iw

RUN echo "" > /var/lib/dhcp/dhcpd.leases
COPY wlanstart.sh /bin/wlanstart.sh

ENTRYPOINT [ "/bin/wlanstart.sh" ]
