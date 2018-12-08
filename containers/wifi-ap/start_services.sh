#!/bin/bash

sysctl -w net.ipv4.ip_forward=1
rfkill block wifi
rfkill unblock wifi
ifup wlan0
iptables-restore < /etc/iptables.ipv4.nat
/opt/replace_wifi_pw.sh
/etc/init.d/dnsmasq start
echo
echo Here are your Docker WiFi credentials:
egrep '(^ssid|pass)' /etc/hostapd/hostapd.conf
echo
/usr/sbin/hostapd -P /run/hostapd.pid /etc/hostapd/hostapd.conf

