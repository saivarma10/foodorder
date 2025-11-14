# Build stage
FROM maven:3.8-openjdk-11 AS build

# Set working directory
WORKDIR /app

# Copy pom.xml and source code
COPY pom.xml .
COPY hello.java .

# Compile the application
RUN mvn clean compile

# Runtime stage
FROM openjdk:11-jre-slim

WORKDIR /app

# Copy compiled classes and dependencies from build stage
COPY --from=build /app/target/classes /app/target/classes
COPY --from=build /app/target/dependency /app/target/dependency
COPY --from=build /root/.m2/repository /root/.m2/repository
COPY pom.xml .

# Install maven (needed for exec:java)
RUN apt-get update && apt-get install -y maven && rm -rf /var/lib/apt/lists/*

# Expose port
EXPOSE 8085

# Run the application
CMD ["mvn", "exec:java"]
