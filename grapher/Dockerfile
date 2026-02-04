# Stage 1: Build the Java application
FROM maven:3.9-eclipse-temurin-21 AS java-build
WORKDIR /src
COPY server/ /src/server/
WORKDIR /src/server
RUN mvn clean package -DskipTests --no-transfer-progress

# Stage 2: Build the TypeScript Analysis Package
FROM node:20-slim AS ts-build
WORKDIR /src
# Copy the analysis package
COPY clients/analysis /src/clients/analysis
WORKDIR /src/clients/analysis

# Install dependencies and compile - using install instead of ci because we modified package.json
RUN npm install && npm run compile

# Stage 3: Final Runtime Image
FROM eclipse-temurin:21-jre-jammy
# Install Node.js in the final JRE image
RUN apt-get update && apt-get install -y curl && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the Java JAR
COPY --from=java-build /src/server/engine/target/server.jar /app/bin/server.jar

# Copy the Compiled TypeScript App
COPY --from=ts-build /src/clients/analysis /app/analysis

# Copy empty config files if they exist or create them
RUN mkdir -p /app/.cobolplugin && \
    echo "[]" > /app/.cobolplugin/pgm_conf.json && \
    echo "[]" > /app/.cobolplugin/proc_grps.json

# Create a symlink for the ccf-cli
RUN cd /app/analysis && npm link

# Set environment variables
ENV JAVA_JAR_PATH=/app/bin/server.jar

# The entrypoint is ccf-cli
ENTRYPOINT ["ccf-cli"]
