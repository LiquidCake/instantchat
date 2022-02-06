FROM nginx

ARG build_ver

#copy static files
RUN mkdir -p /opt/nginx/static/
COPY build/package/nginx/content/static /opt/nginx/static/$build_ver

#copy config
COPY build/package/nginx/conf/nginx.conf /etc/nginx/nginx.conf

#copy ssl cert files
COPY build/package/nginx/conf/ssl/nginx-selfsigned.crt /etc/ssl/certs/nginx-selfsigned.crt
COPY build/package/nginx/conf/ssl/nginx-selfsigned.key /etc/ssl/private/nginx-selfsigned.key

RUN chmod 444 /etc/ssl/certs/nginx-selfsigned.crt
RUN chmod 444 /etc/ssl/private/nginx-selfsigned.key

# Environment Variables
ENV LOG_DIR=/var/log/nginx/

# Create Log Directory
RUN mkdir -p ${LOG_DIR}
RUN chmod 777 -R ${LOG_DIR}

# Declare volumes to mount
VOLUME ${LOG_DIR}

EXPOSE 80 443