# Use Maven image with JDK
FROM maven:3.8-openjdk-11

# Set working directory
WORKDIR /app

# Copy pom.xml and source code
COPY pom.xml .
COPY hello.java .

# Expose port
EXPOSE 8085

# Compile and run the application
CMD ["sh", "-c", "mvn clean compile && mvn exec:java"]
