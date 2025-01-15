FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
RUN mkdir /app
WORKDIR /app

# Copy files
COPY file-srv/out/file-srv-out ./file-srv

COPY file-srv/internal/config/app-config.yml ./app-config.yml

#copy ssl cert files
COPY build/package/nginx/conf/ssl/ssl-bundle.crt /etc/ssl/ssl-bundle.crt
COPY build/package/nginx/conf/ssl/cert.key /etc/ssl/cert.key

RUN chmod 444 /etc/ssl/ssl-bundle.crt
RUN chmod 444 /etc/ssl/cert.key

# Build Args
#is set from docker build cmd line
ARG BUILD_VERSION

# Environment Variables
ENV LOG_DIR=/app/logs
ENV LOG_FILE_NAME=file-srv.log
#Levels: TRACE, INFO, WARN, SEVERE. Overriden in deployment config
ENV LOG_LEVEL=INFO
ENV FILES_DIR=/var/file-srv/uploaded-files

# Create Log Directory
RUN mkdir -p ${LOG_DIR}

# Create uploaded files dir
RUN mkdir -p ${FILES_DIR}/text

# Declare volumes to mount
VOLUME ${LOG_DIR}
VOLUME ${FILES_DIR}

# Expose port 8080 to the outside world
EXPOSE 8080 443

# Command to run the executable
CMD ["/app/file-srv"]

