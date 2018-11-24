---
title: "Docker compose for managing personal server"
tags: [
    "docker",
]
date: "2018-07-31"
categories: [
    "Blog",
]
---

For a long time, I tried to find the most comfortable way to manage a server where I host 
something for myself, some kind of DevOps framework. And I think I found the best way so far. 

Since most of the times things I want to host are useless and I often change my mind, 
there are several requirements for it:

1. **Easy to add/remove new components.**

    Let's say I created a website, and a couple of days later I added a DNS server 
    on the same host, and then I went to Russia, so I also need to host VPN. Later
    I realize I don't really need DNS server, so I want to remove it. 

    Doing these things, I want to do only them. I don't really want to fix website 
    deployment while deploying VPN server, and accidentally remove VPN when stopping DNS server.

2. **No vendor lock.**

    I want to remove, or start a new server with the same configuration on 
    different hosting provider any time I want. Digital Ocean, AWS, Google Cloud, 
    my old computer - whatever.

3. **Automate as much as possible.**

    Deployment, https certificates for subdomains, restarting failed instances, etc. 
    I don't want to care about this at all.

In the beginning, I was setting up Nginx on the host and Let's Encrypt certificates update. 
So to add or remove something new I had to change nginx configuration for a hole server,
what could lead to crashing of all components because they are exposed to the world via 
same Nginx. Also, I need new subdomain certs, because Let's encrypt didn't have 
wildcards back then.

My first attempt to automate it was [this](https://github.com/ngalayko/my_server). Ansible roles
for each component (usually dockerized) and Makefile to execute them. The main problem was
that I had my own CI server running (for some reason). 
That's why if it crashes, you should go and set it up back manually, 
what is always a pain in the ass and takes some time. So a after couple of crashes,
it got boring, I stopped care about it and deleted host.

A couple of weeks ago I started this blog, so I tried to find a way to manage a host
server one more time.

[Here](https://github.com/ngalayko/server) it is.

There are 2 components:

1. [Automated nginx-docker-with-lets-encrypt compose file](https://github.com/ngalayko/server/blob/master/docker-compose.yml)

    What it does is generating nginx configuration based on other containers in the 
    same docker network and taking care of https for them. You can read more in the 
    repository readme file, but basically, it contains three parts: nginx, nginx configuration generator
    and a certificates generator using Let's Encrypt.
    
    That allows us to add new components easily - just add new service to compose file
    or remove one. Also, does not depend on host provider, because can be run on
    pretty much any operating system.

2. [Travis deployment](https://github.com/ngalayko/server/tree/master/.travis)

    It logs in via ssh to your remote server and runs deployment script there. 
    Step-to-step explanation on how to set it up you can find 
    [here](https://gist.github.com/nickbclifford/16c5be884c8a15dca02dca09f65f97bd). 
    Only change I made there - added environment variables export, so it's possible
    to strore secret keys in Travis. 

That's it! Travis [executes](https://github.com/ngalayko/server/blob/master/scripts/update.sh) all compose files 
in the folder, and removes orphan containers.

To add a new component to the system, I need to create a [new folder](https://github.com/ngalayko/server/tree/master/blog)
or [add git submodule](https://github.com/umputun/remark/tree/e278da3cd074b86c5d59359e4f1c615ab6f98b93) with a Dockerized 
app and add a [docker-compose file](https://github.com/ngalayko/server/blob/master/docker-compose.dns.yml)
to run it, following some rules, so nginx container can find it and create routes.

## Links
  * [Travis deployment configuration](https://gist.github.com/nickbclifford/16c5be884c8a15dca02dca09f65f97bd)
  * [Docker Let's Encrypt Nginx proxy companion](https://github.com/JrCs/docker-letsencrypt-nginx-proxy-companion)

