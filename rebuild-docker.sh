#!/bin/sh

#app version
major_version=0
minor_version=9
build_num=0 #dynamic

#stop docker
echo "=== Stopping containers"
docker-compose -f build/package/docker-compose-local-unified.yml down -v

#remove all docker images
docker rmi -f `docker images --format="{{.Repository}}:{{.Tag}}"`


#      Build app


# Increment next build number

build_num_file="./build/build_num"

if [ -e $build_num_file ]; then
  echo "File $build_num_file already exists!"

  while IFS= read -r line;
  do
    build_num=$line
  done < $build_num_file

else
  echo >> $build_num_file
  echo "Created file $build_num_file"
fi

echo ""
echo ""

build_num=$(($build_num + 1))
echo "build num: $build_num"

#write to file
echo $build_num > $build_num_file

export BUILD_VERSION="$major_version.$minor_version.$build_num"
echo "env BUILD_VERSION: $BUILD_VERSION"

# build env
if [ "$1" = "prod" ]; then
  export BUILD_ENV="prod"
else
  export BUILD_ENV="dev"
fi

echo "env BUILD_ENV: " $BUILD_ENV

# Build aux-srv
echo ""
echo "=== Building aux-srv"

#cleanup
rm -f aux-srv/out/aux-srv-out
rm -f aux-srv/out/assistant-out

#import modules
cd aux-srv/ && go mod init instantchat.rooms/instantchat/aux-srv

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILD_VERSION -X main.BuildEnv=$BUILD_ENV" -o out/aux-srv-out ./cmd/aux-srv
#build deploy-assistant
cd ../build/deploy-assistant && go mod init instantchat.rooms/instantchat/deploy-assistant
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILD_VERSION" -o ../../aux-srv/out/assistant-out ./
cd ../../

#build aux-srv docker image
docker build -t aux-srv -f build/package/aux-srv/Dockerfile .
docker tag aux-srv aux-srv:$BUILD_VERSION

# Build file-srv
echo ""
echo "=== Building file-srv"

#cleanup
rm -f file-srv/out/file-srv-out

#import modules
cd file-srv/ && go mod init instantchat.rooms/instantchat/file-srv

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILD_VERSION" -o out/file-srv-out ./cmd/file-srv
cd ..

#build file-srv docker image
docker build -t file-srv -f build/package/file-srv/Dockerfile .
docker tag file-srv file-srv:$BUILD_VERSION


# Build backend
echo ""
echo "=== Building backend"

#cleanup
rm -f backend/out/backend-out
rm -f backend/out/assistant-out

#import modules
cd backend/ && go mod init instantchat.rooms/instantchat/backend

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILD_VERSION" -o out/backend-out ./cmd/backend
#build deploy-assistant
cd ../build/deploy-assistant && go mod init instantchat.rooms/instantchat/deploy-assistant
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-X main.BuildVersion=$BUILD_VERSION" -o ../../backend/out/assistant-out ./
cd ../../

#build backend docker image
docker build -t backend -f build/package/backend/Dockerfile .
docker tag backend backend:$BUILD_VERSION


# Build nginx
echo ""
echo "=== Building nginx"

#build nginx docker image
docker build -t nginx --build-arg build_ver=$BUILD_VERSION -f build/package/nginx/Dockerfile .
docker tag nginx nginx:$BUILD_VERSION


# Run containers
echo ""
echo "=== Starting containers"
docker-compose -f build/package/docker-compose-local-unified.yml up --force-recreate --remove-orphans

echo ""