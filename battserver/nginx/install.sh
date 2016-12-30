#!/bin/bash
docker run --rm --volumes-from nginx -v $(pwd):/currdir ubuntu cp -r /currdir/battleline /var/www/html/;
docker run --rm --volumes-from nginx -v $(pwd):/currdir ubuntu cp /currdir/battleline.conf /etc/nginx/sites-enabled;
docker kill --signal=HUP nginx
