# Amazon EBS only

This rexray plugin currently only supports EBS. It was built as a proof-of-concept.

# How to use it

You can install it on an Amazon EC2 instance with:
```
docker plugin install tiborvass/rexray-plugin EBS_ACCESSKEY=$AWS_ACCESSKEY EBS_SECRETKEY=$AWS_SECRETKEY
```

# How to build it

If you have docker 1.13.0 installed, simply run `./pluginize.sh myname/rexray-plugin`.
