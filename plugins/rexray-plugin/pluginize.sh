set -ex
plugin="$1"
TMPDIR=$(mktemp -d rexray.tmp.XXXXX)
image=$(docker build . -q)
docker create --name rexray "$image"
mkdir -p $TMPDIR/rootfs
docker export -o $TMPDIR/rexray.tar rexray
docker rm -vf rexray
( cd $TMPDIR/rootfs && tar xf ../rexray.tar )
cp config.json $TMPDIR
docker plugin rm "$plugin" || true
docker plugin create "$plugin" "$TMPDIR"
docker plugin push "$plugin" 
rm -rf "$TMPDIR"
