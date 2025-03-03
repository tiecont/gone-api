#!/bin/sh

# Wait for MySQL to be ready (adjust host/port as needed)
echo "Waiting for MySQL to start..."
while ! nc -z gone-mysql 3308; do
  sleep 2
done

echo "MySQL is up - starting application..."
exec java -jar app.jar
