FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
RUN mkdir /app
WORKDIR /app

# Copy files
COPY backend/out/backend-out ./backend
COPY backend/out/assistant-out ./assistant-out

COPY backend/internal/templates/templates_to_compile ./templates_to_compile

COPY backend/internal/config/app-config.yml ./app-config.yml

# Run pre-deploy tool
RUN ./assistant-out >> build.out


# Build Args
#is set from docker build cmd line
ARG BUILD_VERSION

# Environment Variables
ENV LOG_DIR=/app/logs
ENV LOG_FILE_NAME=backend.log
#Levels: TRACE, INFO, WARN, SEVERE. Overriden in deployment config
ENV LOG_LEVEL=INFO

# Create Log Directory
RUN mkdir -p ${LOG_DIR}

# Declare volumes to mount
VOLUME ${LOG_DIR}

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/backend"]

