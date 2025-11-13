# Use Maven image for building
FROM maven:3.8-openjdk-11 AS builder

# Set working directory
WORKDIR /app

# Copy pom.xml
COPY pom.xml .

# Copy source code
COPY hello.java .

# Build the application
RUN mvn clean compile

# Final stage
FROM eclipse-temurin:11-jre-jammy

WORKDIR /app

# Copy compiled classes and dependencies from builder
COPY --from=builder /app/target/classes ./classes
COPY --from=builder /root/.m2/repository /root/.m2/repository
COPY pom.xml .

# Install Maven in runtime (needed to run with exec plugin)
RUN apt-get update && \
    apt-get install -y maven && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Expose port
EXPOSE 8085

# Run the application
CMD ["mvn", "exec:java"]
