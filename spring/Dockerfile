FROM openjdk:23-jdk-slim

WORKDIR /app

# Install required dependencies
RUN apt-get update && apt-get install -y maven && rm -rf /var/lib/apt/lists/*

# Copy project files
COPY pom.xml ./
COPY .mvn .mvn
COPY mvnw ./

# Ensure Maven Wrapper is executable
RUN chmod +x mvnw

# Download dependencies
RUN ./mvnw dependency:go-offline

# Copy the source code
COPY src ./src

# Enable remote debugging (optional)
ENV JAVA_OPTS="-Dspring.devtools.restart.enabled=true -Dspring.devtools.livereload.enabled=true"

# Run the application using Maven
CMD ["mvn", "spring-boot:run"]
