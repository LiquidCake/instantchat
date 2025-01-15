#!/bin/sh

mkdir -p /var/log/nginx

mkdir -p /var/log/aux-srv

mkdir -p /var/file-srv/files
mkdir -p /var/log/file-srv

mkdir -p /var/log/backend

chmod +x ./redeploy.sh
chmod +x ./grafana/setup.sh

#disable web servers if installed
sudo service nginx stop
sudo service apache2 stop
sudo systemctl disable nginx
sudo systemctl disable apache2
