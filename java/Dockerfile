# Use Maven image with JDK
FROM maven:3.8-openjdk-11

# Set working directory
WORKDIR /app

# Copy pom.xml and source code
COPY pom.xml .
COPY hello.java .

# Compile the application during build
RUN mvn clean compile

# Expose port
EXPOSE 8085

# Run the application
CMD ["mvn", "exec:java"]
