---
title: "Internet privacy starter pack"
tags: [
    "privacy",
]
date: "2018-08-20"
categories: [
    "Blog",
]
---

About a month ago I began using some systems to protect my security on the internet a bit more than nothing.

## PiHole

It started when someone shared a link to the [PiHole](https://pi-hole.net/) on Twitter.
It is a self-hosted DNS service that is designed for RaspberryPi for
blocking advertisements and trackers on the DNS level. 

Turns out, about 20% of queries I make are blocked,
and it doesn't hurt daily usage at all. Instead,
now I have nice stats of a website I visit and handy tools to block/whitelist
some of them.

![Pihole](/media/pihole.jpg)

You can install it directly on your router, on every device that you use.
For MacBook and AppleTV I changed DNS settings to use my custom IP address,
for iPhone, I found [this app](https://www.dnsoverride.com/) that can change DNS for all requests, even when you use cellular.
If you want to try it, the address is the same as the IP of this website `167.99.219.223`

## Matomo

![natomo](/media/matomo.jpg)

After enabling PiHole, Google Analytics stopped working, so I found a self-hosted alternative and installed it on this website. 

[Matomo](https://matomo.org/) is pretty good. It has all essential analytics features and respects user privacy. For example, I configured it anonymize last digits of user IP. For example, `167.99.219.223` will be changed to `167.99.219.000` before saving. 

You can check if it tracks you (this is an `iframe`): 

<iframe
	src="https://analytics.galaiko.rocks/index.php?module=CoreAdminHome&action=optOut&language=en&backgroundColor=d3dcda&fontColor=&fontSize=&fontFamily=Helvetica%20Neue"
></iframe>

If you don't want it, you can [configure a browser](https://support.apple.com/kb/PH21416?locale=en_US)
to send "Do not track me" requests to websites. 

## VPN

And, of course, VPN. I hope everyone knows what it is.

This time I use [Libreswan](https://libreswan.org/) and [xl2tpd](https://github.com/xelerance/xl2tpd) setup. 

If you are going to China, try [streisand](https://github.com/StreisandEffect/streisand).
VPN detecting and blocking technology there is the next level,
something simple will be banned within a couple of days for sure. 

## Links
* [PiHole](https://pi-hole.net/)
* [docker-compose file for PiHole](https://github.com/ngalayko/server/blob/master/docker-compose.dns.yml)
* [domains to block](https://firebog.net/)
* [iOS app to change dns](https://www.dnsoverride.com/)
* [Matomo](https://matomo.org/)
* [docker-compose file for Matomo](https://github.com/ngalayko/server/blob/master/docker-compose.analytics.yml)
* [docker-compose file for VPN](https://github.com/ngalayko/server/blob/master/docker-compose.vpn.yml)
* [streisand vpn](https://github.com/StreisandEffect/streisand)
