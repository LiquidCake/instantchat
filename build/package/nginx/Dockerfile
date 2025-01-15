FROM nginx

ARG build_ver

#copy static files
RUN mkdir -p /opt/nginx/static/
COPY build/package/nginx/content/static /opt/nginx/static/$build_ver

#copy config
COPY build/package/nginx/conf/nginx.conf /etc/nginx/nginx.conf

#copy ssl cert files
COPY build/package/nginx/conf/ssl/ssl-bundle.crt /etc/ssl/ssl-bundle.crt
COPY build/package/nginx/conf/ssl/cert.key /etc/ssl/cert.key

RUN chmod 444 /etc/ssl/ssl-bundle.crt
RUN chmod 444 /etc/ssl/cert.key

# Environment Variables
ENV LOG_DIR=/var/log/nginx/

# Create Log Directory
RUN mkdir -p ${LOG_DIR}
RUN chmod 777 -R ${LOG_DIR}

# Declare volumes to mount
VOLUME ${LOG_DIR}

EXPOSE 80 443