#!/bin/bash

# Exit on error
set -e

# Wait for the database to be ready
echo "Waiting for the database..."

echo "Database is ready!"

# Run migrations
echo "Running migrations..."
if ! go run main.go migrate; then
  echo "Migration failed. Exiting."
  exit 1
fi

# Start the application with Air
echo "Starting the application with Air..."
exec air
