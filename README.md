## This is a universally-available instant text chat project
It is meant to be as simple as possible both in terms of usage (for users) and maintenance (for admin)  
Users can create uniquely-named chat rooms that are available by a clear direct URL without any explicit authentication.  
Service is accessible using cross-browser responsive web interface, terminal client or even very simple http requests protocol - to the point when you can use it for simple communication between some IOT devices

Chat is anonymous and meant for 'operative' use - messages are not persisted anywhere (so no chat history) and are lost upon service restart (or after room was inactive for a long time)  
Room will start to erase old messages once count goes beyond some limit (1000 by default)  

###
Service is written in GO (backend) and jquery (frontend)  
GUI is written in JS+Jquery by hand. It is responsive, cross-browser and cross-platform  
Base service monitoring (Grafana/Prometheus) is included  

### Check out author's deployed instance at https://avajoin.net/  

### Screenshots
![large](https://github.com/user-attachments/assets/3940d57f-0b4d-4132-9274-e75376d64886)


####
![small](https://github.com/user-attachments/assets/d33e51ab-20e2-4d6f-a61a-af56fd0b8302)

####

![universal-w98](https://github.com/user-attachments/assets/e83caea9-18e6-4610-84a7-349bbc0b961f)

####

### Architecture and service specifics

Architecture is derived from project initial specifics like all rooms having unique name (to be accessible by a clear human-readable URL), and from desire to have services as autonomous and sustainable as possible (except for backends cluster scaling which currently is manual)

#### Service consists of following components:

`nginx` - for serving static files and proxy-pass to static web pages

`aux-srv` - HTTP server for serving static web pages and load-balancing: it decides which backend is least loaded and places newly created room on that backend to be handled by that backend exclusively. Aux-srv is the single point of knowledge about which rooms are assigned to which backend

(if service is restarted - it will on-demand regain this knowledge from operating backends)

`backend (1..N)` - backend service(s) that serve web-socket (WS) connections. These are not just WS entrypoints, they also hold each room and all its messages in RAM and does all the job (to avoid having global message queue server for storing messages)

So room is placed on some backend instance and all users of that room will connect to exactly that backend (via WS)

Currently there is no way to move room to another backend, if backend is stopped - all its rooms and messages are gone

`file-srv` - simple file server for storing/serving user-made drawings

`grafana` - for monitoring

# Configuration and deployment

There are 3 deployment configurations prepared:
1. local (`Run locally` section) -

```all services (excluding monitoring) are deployed in a local docker via single compose file```

2. multi-node deployment in a 'gateway' mode (`Multi-node deployment in a 'gateway' mode` section) -

```'gateway node' - nginx + aux-srv + file-srv```

```'backend nodes 1, 2, 3' - three backend instances (3 is example, may be configured to any number)```

```'monitoring node' - grafana + prometheus```

`It is possible to deploy nginx, aux-srv, file-srv separately instead of using a 'gateway' node that unifies them, but below examples are given for 'gateway mode' approach`

3. single-node deployment (`Single-node deployment` section) -
```all services are deployed in on-server docker via single compose file```


## Configuration

Check out config files and set your domain, node addresses etc.
```
instantchat/aux-srv/internal/config/app-config.yml
instantchat/backend/internal/config/app-config.yml
```

(for non-local run) set node addresses as prometheus container `extra_hosts` at  
```instantchat/build/deployment/multi-node/monitoring/docker-compose-monitoring.yml``` (for multi-node deployment)  
```instantchat/build/deployment/single-node/docker-compose-single-node.yml``` (for single-node deployment)  

Set domain in nginx.conf - ```instantchat/build/deployment/multi-node/nginx/conf/nginx.conf```  
Also - ONLY if services like aux-srv are deployed separately from nginx, instead of deploying them together as 'gateway' node - override docker's internal service addresses in nginx.conf  

Put a valid ssl certificate for your domain (or self-signed one for local setup) into ```build/package/nginx/conf/ssl```    

Override env vars for `Admin panel` - `CTRL_AUTH_LOGIN, CTRL_AUTH_PASSWD` set in docker-compose file(s) respective to your deployment variant  

After all above steps - take a look at a section about deployment variant you are are going to use (local setup, single-node deployment, multi-node deployment)  

## build project
###### below instructions are both to prepare images for 'prod' deployment and to run locally
After project `configuration` for a particular deployment type (local or multi-node or single-node, see respective sections):  
On a Linux machine (or WSL) with GO compiler and docker installed, run in project root:

Run ```sudo ./rebuild-docker.sh```

This will build all images and start them in local docker (using unified docker file for local deployment - ```instantchat/build/package/docker-compose-local-unified.yml```)

(start / stop manually with ```docker-compose -f instantchat/build/package/docker-compose-local-unified.yml up```)

That's it, images are ready to be deployed to server nodes if you configured project for one of remote deployment variants (see corresponding sections). Stop docker and start uploading images.  
Or if project was configured to run locally - just open your local domain address in browser  

## Run locally
Web UI is https-only, so you will have to use a real or self-signed certificate  

- invent some test domain like e.g. `myinstantchat.org`  

- (unless provided one is not ok) create a self-signed cert and copy into `build/package/nginx/conf/ssl`  
Generate e.g. like this: `sudo openssl req -x509 -nodes -days 999999 -newkey rsa:2048 -keyout cert.key -out ssl-bundle.crt`  

- set domain into all address properties in config files:  
```
instantchat/aux-srv/internal/config/app-config.yml
instantchat/backend/internal/config/app-config.yml  
```

- set `unsecureTestMode: true` in `instantchat/aux-srv/internal/config/app-config.yml` to allow aux-srv to call backend with self-signed cert  

- set domain as extra host for `aux-srv` inside docker-compose file - edit ```instantchat/build/package/docker-compose-local-unified.yml``` and add your local IP (e.g. `192.168.1.100`):
```
  aux-srv:
    extra_hosts:
      - "myinstantchat.org:192.168.1.100"
``` 

- create an OS `hosts` file override for this domain e.g. `192.168.1.100 myinstantchat.org`

- build/launch project (see `build` section), then:  
open `myinstantchat.org` address in browser and accept self-signed cert.  
open backend websocket endpoint address in browser and accept self-signed cert for websockets to work - e.g. `https://myinstantchat.org:12443/ws_entry`. Address/port is the one you set in aux-srv `app-config.yml` as `backendInstances`


## Multi-node deployment in a 'gateway' mode

### configure the project
Take a look at `configuration` section and make necessary changes  

Additionally, for multi-node deployment you would want to add all backend node addresses to monitoring config files - process is described in section `How to add new backend node`

### build the project
Take a look at `build` section - build project after configuration to prepare docker images  

```
For `Multi-node deployment in a 'gateway' mode` - example server node addresses are set across default configs and deployment commands as following:

192.168.1.100 - as IP of 'gateway' node (where nginx+aux-srv are deployed)
192.168.1.101, 192.168.1.102, 192.168.1.103 - as 1st, 2nd and 3rd backend nodes
192.168.1.255 - as monitoring node
```

### 1. Copy files and init environment for all server nodes

#### create OS user on all server nodes
```
sudo adduser instantchat
usermod -aG sudo instantchat
```

#### copy env init scripts to server nodes

(example scripts are made for Ubuntu Linux though doesn't matter really)

###### execute on build machine
```
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.101:/home/instantchat
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.102:/home/instantchat
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.103:/home/instantchat
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.255:/home/instantchat

scp instantchat/build/deployment/multi-node/gateway/init-gateway.sh instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/init-backend.sh instantchat@192.168.1.101:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/init-backend.sh instantchat@192.168.1.102:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/init-backend.sh instantchat@192.168.1.103:/home/instantchat
scp instantchat/build/deployment/multi-node/monitoring/init-monitoring.sh instantchat@192.168.1.255:/home/instantchat

#upload monitoring files
scp -r instantchat/build/deployment/multi-node/monitoring/prometheus/ instantchat@192.168.1.255:/home/instantchat
scp -r instantchat/build/deployment/multi-node/monitoring/grafana/ instantchat@192.168.1.255:/home/instantchat

#upload docker-compose files
scp instantchat/build/deployment/multi-node/gateway/docker-compose-gateway.yml instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/docker-compose-backend-1.yml instantchat@192.168.1.101:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/docker-compose-backend-2.yml instantchat@192.168.1.102:/home/instantchat
scp instantchat/build/deployment/multi-node/backend/docker-compose-backend-3.yml instantchat@192.168.1.103:/home/instantchat
scp instantchat/build/deployment/multi-node/monitoring/docker-compose-monitoring.yml instantchat@192.168.1.255:/home/instantchat
```

#### execute on all server nodes

init env (install docker etc)

```./env-init-ubuntu.sh```

init node-specific configs

#### execute on gateway node

```./init-gateway.sh```

#### execute on all backend nodes

```./init-backend.sh```

#### execute on monitoring node

```./init-monitoring.sh```

### 2. Upload and import docker images we build earlier
#### (needs to be repeated after any new build is ready to deployment)

#### execute on build machine

save images to files
```
sudo docker save -o /tmp/nginx-latest.tar nginx:latest && \
sudo docker save -o /tmp/aux-srv-latest.tar aux-srv:latest && \
sudo docker save -o /tmp/file-srv-latest.tar file-srv:latest && \
sudo docker save -o /tmp/backend-latest.tar backend:latest
```

copy images to gateway node
```
scp /tmp/nginx-latest.tar instantchat@192.168.1.100:/home/instantchat
scp /tmp/aux-srv-latest.tar instantchat@192.168.1.100:/home/instantchat
scp /tmp/file-srv-latest.tar instantchat@192.168.1.100:/home/instantchat
```

copy backend image to all backend nodes
```
scp /tmp/backend-latest.tar instantchat@192.168.1.101:/home/instantchat
scp /tmp/backend-latest.tar instantchat@192.168.1.102:/home/instantchat
scp /tmp/backend-latest.tar instantchat@192.168.1.103:/home/instantchat
```

### load images on server nodes

#### execute on gateway node
```
sudo docker load -i nginx-latest.tar && \
sudo docker load -i aux-srv-latest.tar && \
sudo docker load -i file-srv-latest.tar
```

#### execute on all backend nodes

```sudo docker load -i backend-latest.tar```

### 3. Run docker on remote servers
On each server - run ```docker-compose -f /home/instantchat/docker-compose-{node_specific_file_name} up```

### finalize monitoring setup
Grafana dashboards are not imported automatically, so import manually via GUI.  
Dashboard files for multi-node deployment are located at ```instantchat/build/deployment/multi-node/monitoring/grafana/dashboards```  
(by default Grafana is deployed with default creds `admin/admin` - dont forget to change)  

### How to add new backend node
#### (backends cluster scaling for a Multi-node deployment)

`Currently everything is done manually`

a) prepare compose file docker-compose-backend-XXX.yml (copy existing one)

b) create server, init environment, copy docker file to it, deploy backend image (see above)

c) add new backend server address to aux-srv config - aux-srv/internal/config/app-config.yml

d) add new backend server address to monitoring configs:

- add host to monitoring/docker-compose-monitoring.yml (extra_hosts:)
  
- add host to monitoring/prometheus/prometheus.yml
  
- add host to monitoring/prometheus/sd/targets-node-exporter.yml


e) re-deploy monitoring node (see above)

- copy updated docker-compose-monitoring.yml to monitoring node
  
- re-deploy monitoring services

## Single-node deployment
For single-node deployment (all services are on same remote server) - steps are the same as for `Multi-node deployment in a 'gateway' mode`, except there is only 1 node with a single address (`192.168.1.100` in below examples) - to configure everywhere and to load docker images to.
Take a look at multi-node deployment for more insight  

Single-node configuration steps are listed below  

### configure the project
Take a look at `configuration` section and make necessary changes

### build the project
Take a look at `build` section - build project after configuration to prepare docker images

### prepare env
execute on build machine:

```
scp instantchat/build/deployment/env-init-ubuntu.sh instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/single-node/init-single-node.sh instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/single-node/docker-compose-single-node.yml instantchat@192.168.1.100:/home/instantchat
scp instantchat/build/deployment/single-node/redeploy.sh instantchat@192.168.1.100:/home/instantchat

scp -r instantchat/build/deployment/single-node/prometheus/ instantchat@192.168.1.100:/home/instantchat
scp -r instantchat/build/deployment/single-node/grafana/ instantchat@192.168.1.100:/home/instantchat
```

execute on remote node:
```
./env-init-ubuntu.sh
./init-single-node.sh
```

### upload containers we built earlier
execute on build machine:

```
sudo docker save -o /tmp/nginx-latest.tar nginx:latest && \
sudo docker save -o /tmp/aux-srv-latest.tar aux-srv:latest && \
sudo docker save -o /tmp/file-srv-latest.tar file-srv:latest && \
sudo docker save -o /tmp/backend-latest.tar backend:latest

sudo scp \
/tmp/nginx-latest.tar \
/tmp/aux-srv-latest.tar \
/tmp/file-srv-latest.tar \
/tmp/backend-latest.tar \
instantchat@192.168.1.100:/home/instantchat
```

### import docker images and launch containers
execute on remote node:
```
#WARNING: contains 'docker image prune' which removes unused containers
sudo ./redeploy.sh
```

### finalize monitoring setup
Grafana dashboards are not imported automatically, so import manually via GUI.  
Dashboard files for single-node deployment are located at ```instantchat/build/deployment/single-node/grafana/dashboards```  
(by default Grafana is deployed with default creds `admin/admin` - dont forget to change)  


## Admin panel (kind of)
(wont work on local setup without some nginx tinkering, doesn't like proxying to self-signed cert)
Simple admin panel that allows some runtime operations with backends is available at `https://{domain}/control_page_proxy`
Make sure to change login/password variables `CTRL_AUTH_LOGIN, CTRL_AUTH_PASSWD` in docker-compose file(s)

## Load testing
### Configuration
Configs are set in:
```
instantchat/load-testing/test_client/src/main/java/main/util/Constants.java
instantchat/load-testing/test_client/src/main/java/main/Main.java
```

#### Run
(you will need fresh `java` and `maven` installed)

execute ```./instantchat/load-testing/test_client/run.sh```
