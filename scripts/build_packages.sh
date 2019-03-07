#!/bin/bash
set -x
set -e
BASE=$(dirname $0) # points to scripts directory
CODE_DIR=$(readlink -e "$BASE/../") # project root
BUILD_ROOT=$CODE_DIR/build # should have all binaries already inside
BUILD_PKG=$CODE_DIR/build_pkg # will place packages here
BUILD_TMP=$CODE_DIR/build_tmp # used for temporary data used to construct the packages

# clean any pre-existing cruft
rm -rf $BUILD_TMP $BUILD_PKG

mkdir $BUILD_TMP
mkdir $BUILD_PKG

sudo apt-get install rpm # to be able to make rpms

ARCH="$(uname -m)"
source $CODE_DIR/scripts/version-tag.sh


## debian wheezy ##

BUILD=${BUILD_TMP}/sysvinit
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/metrictank
mkdir -p ${BUILD}/usr/share/metrictank/examples
PKG=${BUILD_PKG}/sysvinit
mkdir -p ${PKG}

cp ${BASE}/config/metrictank-package.ini ${BUILD}/etc/metrictank/metrictank.ini
cp ${BASE}/config/schema-store-cassandra.toml ${BUILD}/etc/metrictank/schema-store-cassandra.toml
cp ${BASE}/config/schema-idx-cassandra.toml ${BUILD}/etc/metrictank/schema-idx-cassandra.toml
cp ${BASE}/config/schema-store-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-store-scylladb.toml
cp ${BASE}/config/schema-idx-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-idx-scylladb.toml
cp ${BASE}/config/index-rules.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-schemas.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-aggregation.conf ${BUILD}/etc/metrictank/
cp ${BUILD_ROOT}/{metrictank,mt-*} ${BUILD}/usr/bin/

PACKAGE_NAME="${PKG}/metrictank-${version_raw}_${ARCH}.deb"
fpm -s dir -t deb \
  -v ${version_raw} -n metrictank -a ${ARCH} --description "metrictank, the gorilla-inspired timeseries database backend for graphite" \
  --deb-init ${BASE}/config/sysvinit/init.d/metrictank \
  --deb-default ${BASE}/config/sysvinit/default/metrictank \
  --replaces metric-tank --provides metric-tank \
  --conflicts metric-tank \
  --config-files /etc/metrictank/ \
  -C ${BUILD} -p ${PACKAGE_NAME} .


## ubuntu 14.04 ##

BUILD=${BUILD_TMP}/upstart
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/init
mkdir -p ${BUILD}/etc/metrictank
mkdir -p ${BUILD}/usr/share/metrictank/examples
PKG=${BUILD_PKG}/upstart
mkdir -p ${PKG}

cp ${BASE}/config/metrictank-package.ini ${BUILD}/etc/metrictank/metrictank.ini
cp ${BASE}/config/schema-store-cassandra.toml ${BUILD}/etc/metrictank/schema-store-cassandra.toml
cp ${BASE}/config/schema-idx-cassandra.toml ${BUILD}/etc/metrictank/schema-idx-cassandra.toml
cp ${BASE}/config/schema-store-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-store-scylladb.toml
cp ${BASE}/config/schema-idx-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-idx-scylladb.toml
cp ${BASE}/config/index-rules.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-schemas.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-aggregation.conf ${BUILD}/etc/metrictank/
cp ${BUILD_ROOT}/{metrictank,mt-*} ${BUILD}/usr/bin/

PACKAGE_NAME="${PKG}/metrictank-${version_raw}_${ARCH}.deb"
fpm -s dir -t deb \
  -v ${version_raw} -n metrictank -a ${ARCH} --description "metrictank, the gorilla-inspired timeseries database backend for graphite" \
  --deb-upstart ${BASE}/config/upstart/metrictank \
  --replaces metric-tank --provides metric-tank \
  --conflicts metric-tank \
  --config-files /etc/metrictank/ \
  -C ${BUILD} -p ${PACKAGE_NAME} .


## ubuntu 16.04, Debian 8, CentOS 7 ##

BUILD=${BUILD_TMP}/systemd
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/lib/systemd/system/
mkdir -p ${BUILD}/etc/metrictank
mkdir -p ${BUILD}/usr/share/metrictank/examples
mkdir -p ${BUILD}/var/run/metrictank
PKG=${BUILD_PKG}/systemd
mkdir -p ${PKG}

cp ${BASE}/config/metrictank-package.ini ${BUILD}/etc/metrictank/metrictank.ini
cp ${BASE}/config/schema-store-cassandra.toml ${BUILD}/etc/metrictank/schema-store-cassandra.toml
cp ${BASE}/config/schema-idx-cassandra.toml ${BUILD}/etc/metrictank/schema-idx-cassandra.toml
cp ${BASE}/config/schema-store-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-store-scylladb.toml
cp ${BASE}/config/schema-idx-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-idx-scylladb.toml
cp ${BASE}/config/index-rules.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-schemas.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-aggregation.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/systemd/metrictank.service $BUILD/lib/systemd/system/
cp ${BUILD_ROOT}/{metrictank,mt-*} ${BUILD}/usr/bin/

PACKAGE_NAME="${PKG}/metrictank-${version_raw}_${ARCH}.deb"
fpm -s dir -t deb \
  -v ${version_raw} -n metrictank -a ${ARCH} --description "metrictank, the gorilla-inspired timeseries database backend for graphite" \
  --config-files /etc/metrictank/ \
  -m "Raintank Inc. <hello@grafana.com>" --vendor "grafana.com" \
  --license "Apache2.0" -C ${BUILD} -p ${PACKAGE_NAME} .


## centos 7 ##

BUILD=${BUILD_TMP}/systemd-centos7
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/lib/systemd/system/
mkdir -p ${BUILD}/etc/metrictank
mkdir -p ${BUILD}/usr/share/metrictank/examples
mkdir -p ${BUILD}/var/run/metrictank
PKG=${BUILD_PKG}/systemd-centos7
mkdir -p ${PKG}

cp ${BASE}/config/metrictank-package.ini ${BUILD}/etc/metrictank/metrictank.ini
cp ${BASE}/config/schema-store-cassandra.toml ${BUILD}/etc/metrictank/schema-store-cassandra.toml
cp ${BASE}/config/schema-idx-cassandra.toml ${BUILD}/etc/metrictank/schema-idx-cassandra.toml
cp ${BASE}/config/schema-store-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-store-scylladb.toml
cp ${BASE}/config/schema-idx-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-idx-scylladb.toml
cp ${BASE}/config/index-rules.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-schemas.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-aggregation.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/systemd/metrictank.service $BUILD/lib/systemd/system/
cp ${BUILD_ROOT}/{metrictank,mt-*} ${BUILD}/usr/bin/

PACKAGE_NAME="${PKG}/metrictank-${version_raw}.el7.${ARCH}.rpm"
fpm -s dir -t rpm \
  -v ${version_raw} -n metrictank -a ${ARCH} --description "metrictank, the gorilla-inspired timeseries database backend for graphite" \
  --config-files /etc/metrictank/ \
  -m "Raintank Inc. <hello@grafana.com>" --vendor "grafana.com" \
  --license "Apache2.0" -C ${BUILD} -p ${PACKAGE_NAME} .


## CentOS 6 ##

BUILD=${BUILD_TMP}/upstart-0.6.5
mkdir -p ${BUILD}/usr/bin
mkdir -p ${BUILD}/etc/init
mkdir -p ${BUILD}/etc/metrictank
mkdir -p ${BUILD}/usr/share/metrictank/examples
PKG=${BUILD_PKG}/upstart-0.6.5
mkdir -p ${PKG}

cp ${BASE}/config/metrictank-package.ini ${BUILD}/etc/metrictank/metrictank.ini
cp ${BASE}/config/schema-store-cassandra.toml ${BUILD}/etc/metrictank/schema-store-cassandra.toml
cp ${BASE}/config/schema-idx-cassandra.toml ${BUILD}/etc/metrictank/schema-idx-cassandra.toml
cp ${BASE}/config/schema-store-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-store-scylladb.toml
cp ${BASE}/config/schema-idx-scylladb.toml ${BUILD}/usr/share/metrictank/examples/schema-idx-scylladb.toml
cp ${BASE}/config/index-rules.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-schemas.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/storage-aggregation.conf ${BUILD}/etc/metrictank/
cp ${BASE}/config/upstart-0.6.5/metrictank.conf $BUILD/etc/init
cp ${BUILD_ROOT}/{metrictank,mt-*} ${BUILD}/usr/bin/

PACKAGE_NAME="${PKG}/metrictank-${version_raw}.el6.${ARCH}.rpm"
fpm -s dir -t rpm \
  -v ${version_raw} -n metrictank -a ${ARCH} --description "metrictank, the gorilla-inspired timeseries database backend for graphite" \
  --replaces metric-tank --provides metric-tank \
  --conflicts metric-tank \
  --config-files /etc/metrictank/ \
  -C ${BUILD} -p ${PACKAGE_NAME} .
