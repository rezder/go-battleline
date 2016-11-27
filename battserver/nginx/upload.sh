#!/bin/bash
MYDIR="$(dirname "$(realpath "$0")")"
tar -cf nginxup.tar battleline.conf install.sh
cd html
tar --append --file=$MYDIR/nginxup.tar battleline
cd $MYDIR
gzip nginxup.tar # To extract: $ tar -xzf nginxup.tar.gz
scp  nginxup.tar.gz rho@rezder.com:/home/rho/upload/battleline
rm nginxup.tar.gz
