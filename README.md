### This is an instant text chat project

It is meant to be as simple as possible both in terms of usage (for users) and maintenance (for admin)

Users can create unique-named chat rooms that are available by a clear direct URL, without registration.

Chat is anonymous and 'operative' - messages are not persisted anywhere (so no chat history) and are lost upon service restart (or after room was inactive for a long time).

Room will have only up to X (1000) messages and after that begin to erase old messages

###

Service is written in GO (backend) and jquery (frontend)

GUI is adaptive, crossbrowser and crossplatform (with EXCEPTION to IOS currently as i don't have a devices - tap behaviour seems to be not captured by js correctly)

Includes base monitoring (Grafana/Prometheus)


![screenshot1](https://user-images.githubusercontent.com/9273621/152684403-41758f22-037e-43b7-854c-461e0ed463ed.png)

####

![screenshot2](https://user-images.githubusercontent.com/9273621/152684406-01c371f0-2e0d-4610-86cb-40dc7f8d5b20.png)

### Architecture and service specifics

Architecture is derived from project initial specifics like all rooms having unique name (to be accessible by a clear human-readable URL), and from desire to have backends to be as autonomous and sustainable as possible (except for backends cluster scaling which is manual)

#### Service consists of following components:

`nginx` - for serving static files and proxy-pass to static web pages

`aux-srv` - HTTP server for serving static web pages and load-balancing: it decides which backend is least loaded and places newly created room on that backend to be handled by that backend exclusively. Aux-srv is the single point of knowledge about which rooms are assigned to which backend

(if service is restarted - it will on-demand regain this knowledge from operating backends)

`backend (1..N)` - backend service(s) that serve web-socket (WS) connections. These are not just WS entrypoints, they also hold each room and all its messages in RAM and does all the job (to avoid having global message queue server for storing messages)

There is no way to move room to another backend, if backend is stopped - all its rooms and messages are gone

So room is placed on some backend instance and all users of that room will connect to exactly that backend (via WS)

`file-srv` - simple file server for storing/serving user-made drawings

`grafana` - for monitoring

# Configuration and deployment

There are 2 deployment configurations prepared:
1. local (`'Run locally'` section) -

```all services are deployed in a local docker via single compose file```

2. server deployment in a 'gateway' mode (`Server deployment in a 'gateway' mode` section) -

```'gateway node' - nginx + aux-srv + file-srv```

```'backend nodes 1, 2, 3' - three backend instances (may be configured to any number)```

```'monitoring node' - grafana```


```
example server node addresses are set across default configs and deployment commands as following:

192.168.1.100 - as IP of 'gateway' node (where nginx+aux-srv are deployed)
192.168.1.101, 192.168.1.102, 192.168.1.103 - as 1st, 2nd and 3rd backend nodes
192.168.1.255 - as monitoring node
```

It is possible to deploy all services separately instead of using 'gateway' node that unifies several, but below examples are given for 'gateway mode' approach

## Configuration:

Set node addresses, domain etc.

config files:
```
/instantchat/aux-srv/internal/config/app-config.yml
/instantchat/backend/internal/config/app-config.yml
```

set 'extra_hosts' at

```/instantchat/build/nodes/monitoring/docker-compose-monitoring.yml```

(only if services like aux-srv are deployed separately from nginx, instead of deploying them together as 'gateway' node) - set service addresses in nginx.conf

```/instantchat/build/package/nginx/conf/nginx.conf```


## Build:
(run after configuration, i.e. node addresses set etc.)

on a Linux machine (or WSL) with GO compiler and docker installed, run in project root:

```./rebuild-docker.sh [dev, prod]```

This will build all images and start them in local docker (using unified docker file for local deployment - instantchat/build/package/docker-compose-local-unified.yml)

Thats it, images are ready to be deployed to server nodes and also service is up locally

### Run locally:
set local machine IP and local docker backend addresses in config files:
```
set 'backend1', 'backend2', 'backend3' to 'backendInstances' in /instantchat/aux-srv/internal/config/app-config.yml
set local machine IP to 'allowedOrigins' in /instantchat/backend/internal/config/app-config.yml
```

Build (with ```./rebuild-docker.sh dev```) and start project 

(will start after build or run ```docker-compose -f /instantchat/build/package/docker-compose-local-unified.yml up```)


## Server deployment in a 'gateway' mode

### 1. Init environment for all server nodes

copy env init scripts to server nodes

(example scripts are made for Ubuntu Linux though doesn't matter really)

#### execute on build node
```
scp /instantchat/build/nodes/env-init-ubuntu.sh instantchat@192.168.1.100:/home/instantchat
scp /instantchat/build/nodes/env-init-ubuntu.sh instantchat@192.168.1.101:/home/instantchat
scp /instantchat/build/nodes/env-init-ubuntu.sh instantchat@192.168.1.102:/home/instantchat
scp /instantchat/build/nodes/env-init-ubuntu.sh instantchat@192.168.1.103:/home/instantchat
scp /instantchat/build/nodes/env-init-ubuntu.sh instantchat@192.168.1.255:/home/instantchat

scp /instantchat/build/nodes/gateway/init-gateway.sh instantchat@192.168.1.100:/home/instantchat

scp /instantchat/build/nodes/backend/init-backend.sh instantchat@192.168.1.101:/home/instantchat
scp /instantchat/build/nodes/backend/init-backend.sh instantchat@192.168.1.102:/home/instantchat
scp /instantchat/build/nodes/backend/init-backend.sh instantchat@192.168.1.103:/home/instantchat
```

#### execute on all server nodes

add OS user
```
sudo adduser instantchat
usermod -aG sudo instantchat
```

init env (install docker etc)

```./env-init-ubuntu.sh```

init node-specific configs

#### execute on gateway node

```./init-gateway.sh```

#### execute on all backend nodes

```./init-backend.sh```


### 2. Copy compose files to nodes

#### execute on build node
```
scp /instantchat/build/nodes/gateway/docker-compose-gateway.yml instantchat@192.168.1.100:/home/instantchat

scp /instantchat/build/nodes/backend/docker-compose-backend-1.yml instantchat@192.168.1.101:/home/instantchat
scp /instantchat/build/nodes/backend/docker-compose-backend-2.yml instantchat@192.168.1.102:/home/instantchat
scp /instantchat/build/nodes/backend/docker-compose-backend-3.yml instantchat@192.168.1.103:/home/instantchat
```

### 3. Deploy images 
#### (needs to be repeated after any new build is ready to deployment)

After proper node addresses are `configured` and images are `built` on build node:

#### execute on build node

save images to files
```
docker save -o /tmp/nginx-latest.tar nginx:latest
docker save -o /tmp/aux-srv-latest.tar aux-srv:latest
docker save -o /tmp/file-srv-latest.tar file-srv:latest
docker save -o /tmp/backend-latest.tar backend:latest
```

copy images to gateway node
```
scp /tmp/nginx-latest.tar instantchat@192.168.1.100:/home/instantchat
scp /tmp/aux-srv-latest.tar instantchat@192.168.1.100:/home/instantchat
scp /tmp/file-srv-latest.tar instantchat@192.168.1.100:/home/instantchat
```

copy images to all backend nodes
```
scp /tmp/backend-latest.tar instantchat@192.168.1.101:/home/instantchat
scp /tmp/backend-latest.tar instantchat@192.168.1.102:/home/instantchat
scp /tmp/backend-latest.tar instantchat@192.168.1.103:/home/instantchat
```

(optionally) cleanup build node
```
rm /tmp/nginx-latest.tar
rm /tmp/aux-srv-latest.tar
rm /tmp/file-srv-latest.tar
rm /tmp/backend-latest.tar
```

### load images on server nodes

#### execute on gateway node
```
sudo docker load -i nginx-latest.tar
sudo docker load -i aux-srv-latest.tar
sudo docker load -i file-srv-latest.tar
```

#### execute on all backend nodes

```sudo docker load -i backend-latest.tar```


### 4. Deploy monitoring files

#### execute on build node
```
scp /instantchat/build/nodes/monitoring/docker-compose-monitoring.yml instantchat@192.168.1.255:/home/instantchat

scp -r /instantchat/build/nodes/monitoring/prometheus/ instantchat@192.168.1.255:/home/instantchat
scp -r /instantchat/build/nodes/monitoring/grafana/ instantchat@192.168.1.255:/home/instantchat
```

Note: grafana dashboards are not imported automatically, so import manually via GUI. 

Dashboard files are at ````/instantchat/build/nodes/monitoring/grafana/dashboards````


## How to add new backend node:
#### (backends cluster scaling)

a) prepare compose file docker-compose-backend-XXX.yml (copy existing one)

b) create server, init environment, copy docker file to it, deploy backend image (see above)

c) add new backend server address to aux-srv config - aux-srv/internal/config/app-config.yml

d) add new backend server address to monitoring configs:

- add host to monitoring/docker-compose-monitoring.yml (extra_hosts:)
  
- add host to monitoring/prometheus/prometheus.yml
  
- add host to monitoring/prometheus/sd/targets-node-exporter.yml


e) re-deploy monitoring node (see above)

- copy updated docker-compose-monitoring.yml to monitoring node
  
- deploy updated monitoring files to monitoring node


## Load testing
### Configuration
Configs are set in:
```
/instantchat/load-testing/test_client/src/main/java/main/util/Constants.java
/instantchat/load-testing/test_client/src/main/java/main/Main.java
```

#### Run
(you will need fresh `java` and `maven` installed)

execute ```./instantchat/load-testing/test_client/run.sh```
