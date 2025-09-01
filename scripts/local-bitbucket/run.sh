#!/bin/bash

set -xv

# Start Bitbucket Server Docker container
# https://hub.docker.com/r/atlassian/bitbucket-server

# make sure 127.0.0.1 bitbucket.example.com is added to your /etc/hosts

BITBUCKET_IMAGE_TAG=$1
BITBUCKET_HOME=$2
BITBUCKET_HOST=$3
PERSIST_BITBUCKET_LOCALLY=$4

echo ""
echo "Starting fresh install of Bitbucket Server, using tag: ${BITBUCKET_IMAGE_TAG}"

if [ "${GHORG_GHA_CI:-}" == "true" ]; then
  GHORG_SSH_PORT=2224
else
  GHORG_SSH_PORT=22
fi

# Pre-configure Bitbucket Server with internal H2 database
echo "Starting Bitbucket Server with pre-configured internal H2 database..."

# Create directory for persistence if needed
if [ "${PERSIST_BITBUCKET_LOCALLY}" == "true" ]; then
    BITBUCKET_DATA_DIR="${BITBUCKET_HOME}"
else
    BITBUCKET_DATA_DIR=$(mktemp -d)
    echo "Using temporary data directory: ${BITBUCKET_DATA_DIR}"
fi

if [ "${PERSIST_BITBUCKET_LOCALLY}" == "true" ];then
  echo "Removing any previous install at path: ${BITBUCKET_HOME}"
  echo ""
  rm -rf "${BITBUCKET_HOME}"
  mkdir -p "${BITBUCKET_HOME}"
fi

# Create shared directory and pre-configure database
mkdir -p "${BITBUCKET_DATA_DIR}/shared"

# Note: Bitbucket Server requires manual setup due to licensing requirements
# The container will start and be ready for manual configuration

# Also create server.xml configuration to ensure proper startup
mkdir -p "${BITBUCKET_DATA_DIR}/shared/config"
cat > "${BITBUCKET_DATA_DIR}/shared/config/server.xml" << EOF
<?xml version="1.0" encoding="utf-8"?>
<Server port="8006" shutdown="SHUTDOWN">
    <Service name="Catalina">
        <Connector port="7990" protocol="HTTP/1.1"
                   connectionTimeout="20000"
                   useBodyEncodingForURI="true"
                   compression="on"
                   compressionMinSize="2048"
                   noCompressionUserAgents="gozilla, traviata"
                   compressableMimeType="text/html,text/xml,text/plain,text/css,application/json,application/javascript,text/javascript"
                   maxThreads="48"
                   minSpareThreads="10"
                   enableLookups="false"
                   acceptCount="10"
                   secure="false"
                   scheme="http"
                   proxyName="${BITBUCKET_HOST}"
                   proxyPort="7990"/>
        <Engine name="Catalina" defaultHost="localhost">
            <Host name="localhost" appBase="webapps" unpackWARs="true" autoDeploy="true">
                <Context path="" docBase="\${bitbucket.home}/atlassian-bitbucket" reloadable="false" useHttpOnly="true">
                    <Manager pathname=""/>
                </Context>
            </Host>
        </Engine>
    </Service>
</Server>
EOF

echo "✅ Bitbucket Server data directory prepared"

# Start Bitbucket Server with PostgreSQL using Docker Compose
echo "Starting Bitbucket Server with PostgreSQL database..."

# Get the directory containing docker-compose.yml
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set environment variables for docker-compose
export BITBUCKET_HOST="${BITBUCKET_HOST}"
export BITBUCKET_IMAGE_TAG="${BITBUCKET_IMAGE_TAG}"
export BITBUCKET_DATA_DIR="${BITBUCKET_DATA_DIR}"

# Update docker-compose.yml with persistent data directory if needed
if [ "${PERSIST_BITBUCKET_LOCALLY}" == "true" ]; then
  sed -i.bak "s|bitbucket_data:/var/atlassian/application-data/bitbucket|${BITBUCKET_DATA_DIR}:/var/atlassian/application-data/bitbucket|" "${SCRIPT_DIR}/docker-compose.yml"
fi

cd "${SCRIPT_DIR}"
docker-compose up -d

# Wait for services to be healthy
echo "Waiting for PostgreSQL to be ready..."
docker-compose exec postgres pg_isready -U bitbucket -d bitbucket || sleep 10

echo "Waiting for Bitbucket Server to be ready..."

# Clean up temp directory if not persisting
if [ "${PERSIST_BITBUCKET_LOCALLY}" != "true" ]; then
  echo "Note: Using temporary data directory: ${BITBUCKET_DATA_DIR}"
fi

echo ""
