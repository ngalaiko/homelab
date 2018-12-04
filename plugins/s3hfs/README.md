# Docker volume plugin for sshFS

This plugin allows you to mount remote folder using sshfs in your container easily.

[![TravisCI](https://travis-ci.org/vieux/docker-volume-sshfs.svg)](https://travis-ci.org/vieux/docker-volume-sshfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/vieux/docker-volume-sshfs)](https://goreportcard.com/report/github.com/vieux/docker-volume-sshfs)

## Usage

### Using a password

1 - Install the plugin

```
$ docker plugin install vieux/sshfs

# or to enable debug 
docker plugin install vieux/sshfs DEBUG=1

# or to change where plugin state is stored
docker plugin install vieux/sshfs state.source=<any_folder>
```

2 - Create a volume

> Make sure the ***source path on the ssh server was exists***.
> 
> Or you'll be failed while use/mount the volume.

```
$ docker volume create -d vieux/sshfs -o sshcmd=<user@host:path> -o password=<password> [-o port=<port>] [-o <any_sshfs_-o_option> ] sshvolume
sshvolume
$ docker volume ls
DRIVER              VOLUME NAME
local               2d75de358a70ba469ac968ee852efd4234b9118b7722ee26a1c5a90dcaea6751
local               842a765a9bb11e234642c933b3dfc702dee32b73e0cf7305239436a145b89017
local               9d72c664cbd20512d4e3d5bb9b39ed11e4a632c386447461d48ed84731e44034
local               be9632386a2d396d438c9707e261f86fd9f5e72a7319417901d84041c8f14a4d
local               e1496dfe4fa27b39121e4383d1b16a0a7510f0de89f05b336aab3c0deb4dda0e
vieux/sshfs         sshvolume
```

3 - Use the volume

```
$ docker run -it -v sshvolume:<path> busybox ls <path>
```

### Using an ssh key

1 - Install the plugin

```
$ docker plugin install vieux/sshfs sshkey.source=/home/<user>/.ssh/

# or to enable debug 
docker plugin install vieux/sshfs DEBUG=1 sshkey.source=/home/<user>/.ssh/

# or to change where plugin state is stored
docker plugin install vieux/sshfs state.source=<any_folder> sshkey.source=/home/<user>/.ssh/
```

2 - Create a volume

> Make sure the ***source path on the ssh server was exists***.
> 
> Or you'll be failed while use/mount the volume.

```
$ docker volume create -d vieux/sshfs -o sshcmd=<user@host:path> [-o IdentityFile=/root/.ssh/<key>] [-o port=<port>] [-o <any_sshfs_-o_option> ] sshvolume
sshvolume
$ docker volume ls
DRIVER              VOLUME NAME
local               2d75de358a70ba469ac968ee852efd4234b9118b7722ee26a1c5a90dcaea6751
local               842a765a9bb11e234642c933b3dfc702dee32b73e0cf7305239436a145b89017
local               9d72c664cbd20512d4e3d5bb9b39ed11e4a632c386447461d48ed84731e44034
local               be9632386a2d396d438c9707e261f86fd9f5e72a7319417901d84041c8f14a4d
local               e1496dfe4fa27b39121e4383d1b16a0a7510f0de89f05b336aab3c0deb4dda0e
vieux/sshfs         sshvolume
```

3 - Use the volume

```
$ docker run -it -v sshvolume:<path> busybox ls <path>
```

## LICENSE

MIT
