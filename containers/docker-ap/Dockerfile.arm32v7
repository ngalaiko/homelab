FROM pipill/armhf-alpine
COPY create_ap /usr/bin/
COPY create_ap.conf /etc/
RUN [ "cross-build-start" ]
RUN apk --update --no-cache add coreutils bash util-linux procps hostapd iproute2 iw dnsmasq iptables 
RUN [ "cross-build-end" ]
ENV AP_WIFI_IFACE=wlan0 AP_INTERNET_IFACE=eth0 AP_SSID=MyAccessPoint AP_PASSPHRASE=MyPassPhrase 

COPY run_create_ap /usr/bin/
CMD ["/usr/bin/run_create_ap"]
