# ---- Build Stage ----
FROM openjdk:23-jdk-slim AS build

WORKDIR /app

# Install dependencies
RUN apt-get update && apt-get install -y curl unzip

# Manually install Gradle (to avoid timeout issues)
ENV GRADLE_VERSION=8.12.1
RUN curl -fsSL https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip -o gradle.zip \
    && unzip gradle.zip -d /opt/ \
    && rm gradle.zip \
    && ln -s /opt/gradle-${GRADLE_VERSION}/bin/gradle /usr/bin/gradle

# Copy project files
COPY . .

# Grant execution permission to Gradle wrapper (if needed)
RUN chmod +x ./gradlew

# Run Gradle build with bootJar (using installed Gradle)
RUN gradle clean bootJar --no-daemon -x test

# ---- Runtime Stage ----
FROM openjdk:23-jdk-slim

WORKDIR /app

# Copy the built JAR file
COPY --from=build /app/build/libs/*.jar app.jar

# Expose the application port
EXPOSE 2001

# Run the application
ENTRYPOINT ["java", "-jar", "app.jar"]
