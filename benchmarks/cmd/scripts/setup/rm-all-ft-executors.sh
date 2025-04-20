#!/bin/bash

# Script to find and forcefully remove Docker containers matching 'ft-executor-*' by name

echo "Looking for containers named 'ft-executor-*' (including stopped)..."

# Get Names of all containers (running and stopped) matching the name filter
# Using {{.Names}} instead of {{.ID}}
mapfile -t container_names < <(docker ps --all --filter "name=ft-executor-*" --format "{{.Names}}" --no-trunc)

# Check if any containers were found
if [ ${#container_names[@]} -eq 0 ]; then
  echo "No containers found matching 'ft-executor-*'."
  exit 0 # Exit cleanly as there's nothing to do
fi

echo "Found the following container names to be removed:"
# List names that will be removed
printf " - %s\n" "${container_names[@]}"

echo "Starting container removal..."

# Loop through the array of container names
for container_name in "${container_names[@]}"; do
  # Force remove the container by name (-f allows removal of running containers)
  echo "Removing container: $container_name"
  # docker rm works with names as well as IDs
  docker rm -f "$container_name"

  # Optional: Check if the removal command failed
  if [ $? -ne 0 ]; then
    echo "Warning: Failed to remove container $container_name." >&2
    # Add error handling logic here if needed (e.g., exit 1)
  fi
done

echo "Container removal process completed."

exit 0
