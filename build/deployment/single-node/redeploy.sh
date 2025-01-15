#!/bin/sh
sudo docker-compose -f ./docker-compose-single-node.yml down

sudo docker load -i ./nginx-latest.tar && \
        sudo docker load -i ./aux-srv-latest.tar && \
        sudo docker load -i ./file-srv-latest.tar && \
        sudo docker load -i ./backend-latest.tar

echo y | sudo docker image prune

sudo docker-compose -f ./docker-compose-single-node.yml up
