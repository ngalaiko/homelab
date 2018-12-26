# Docker container stack: hostap + dhcp server 

This container starts wireless access point (hostap) and dhcp server in docker
container. It supports both host networking and network interface reattaching
to container network namespace modes (host and guest).

## Requirements

On the host system install required wifi drivers, then make sure your wifi adapter
supports AP mode:

```
# iw list
...
        Supported interface modes:
                 * IBSS
                 * managed
                 * AP
                 * AP/VLAN
                 * WDS
                 * monitor
                 * mesh point
...
```

Set country regulations, for example, for Spain set:

```
# iw reg set ES
country ES: DFS-ETSI
        (2400 - 2483 @ 40), (N/A, 20), (N/A)
        (5150 - 5250 @ 80), (N/A, 23), (N/A), NO-OUTDOOR
        (5250 - 5350 @ 80), (N/A, 20), (0 ms), NO-OUTDOOR, DFS
        (5470 - 5725 @ 160), (N/A, 26), (0 ms), DFS
        (57000 - 66000 @ 2160), (N/A, 40), (N/A)
```

## Build / run

* Using host networking:

```
sudo docker run -d -t -e INTERFACE=wlan0 --net host --privileged offlinehacker/docker-ap
```

* Using network interface reattaching:

```
sudo docker run -d -t -e INTERFACE=wlan0 -v /var/run/docker.sock:/var/run/docker.sock --privileged offlinehacker/docker-ap
```

This mode requires access to docker socket, so it can run a short lived
container that reattaches network interface to network namespace of this
container. It also renames wifi interface to **wlan0**, so you get
deterministic networking environment. This mode can be usefull for example for
pentesting, where can you use docker compose to run other wifi hacking tools
and have deterministic environment with wifi interface.

## Environment variables

* **INTERFACE**: name of the interface to use for wifi access point (default: wlan0)
* **OUTGOING**: outgoing network interface (default: eth0)
* **CHANNEL**: WIFI channel (default: 6)
* **SUBNET**: Network subnet (default: 192.168.254.0)
* **AP_ADDR**: Access point address (default: 192.168.254.1)
* **SSID**: Access point SSID (default: docker-ap)
* **WPA_PASSPHRASE**: WPA password (default: passw0rd)
* **HW_MODE**: WIFI mode to use (default: g) 
* **DRIVER**: WIFI driver to use (default: nl80211)
* **HT_CAPAB**: WIFI HT capabilities for 802.11n (default: [HT40-][SHORT-GI-20][SHORT-GI-40]) 
* **MODE**: Mode to run in guest/host (default: host)

## License

MIT

## Author

Jaka Hudoklin <jakahudoklin@gmail.com>

Thanks to https://github.com/sdelrio/rpi-hostap for providing original
implementation.
