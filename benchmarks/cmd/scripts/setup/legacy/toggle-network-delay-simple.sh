#!/bin/bash

# --- Configuration ---
DEVICE="eth0"  # Network interface name, adjust if necessary (e.g., enp0s3, ens192)
DELAY="3ms"    # Delay value to set

# --- Script Logic ---

# Check if an argument was provided
if [ -z "$1" ]; then
  echo "Error: No argument provided." >&2
  echo "Usage: $0 [on|off]" >&2
  exit 1
fi

# Process the argument
case "$1" in
  on)
    echo "Applying ${DELAY} delay to ${DEVICE}..."
    # Execute tc command to add delay using sudo
    # Note: 'add' may fail if a root qdisc already exists on ${DEVICE}.
    # Consider using 'replace' instead of 'add' to overwrite existing settings.
    sudo tc qdisc add dev ${DEVICE} root netem delay ${DELAY}

    # Check the exit status of the last command
    if [ $? -eq 0 ]; then
      echo "Successfully added ${DELAY} delay to ${DEVICE}."
      # You can verify the current qdisc with:
      # tc qdisc show dev ${DEVICE}
    else
      echo "Error: Failed to add delay to ${DEVICE}." >&2
      echo "Possible reasons:" >&2
      echo "  - Device ${DEVICE} does not exist or name is incorrect." >&2
      echo "  - A root qdisc already exists on ${DEVICE}. Remove it first or use 'replace'." >&2
      echo "  - Insufficient privileges to run the sudo command." >&2
      exit 1
    fi
    ;;

  off)
    echo "Removing netem delay from ${DEVICE}..."
    # Execute tc command to delete netem qdisc using sudo
    sudo tc qdisc del dev ${DEVICE} root netem

    # Check the exit status of the last command
    # 'tc qdisc del' returns non-zero (usually 2) if the specified qdisc is not found.
    status=$?
    if [ $status -eq 0 ]; then
      echo "Successfully removed netem qdisc from ${DEVICE}."
    elif [ $status -eq 2 ]; then
       # Exit status 2 usually means "RTNETLINK answers: No such file or directory"
       # This implies the qdisc was already gone, which achieves the desired state.
       echo "Netem qdisc not found on ${DEVICE} (already removed or never added)."
       # Exit successfully in this case too.
       exit 0
    else
      echo "Error: Failed to remove netem qdisc from ${DEVICE} (exit status: $status)." >&2
      exit 1
    fi
    ;;

  *)
    # Handle invalid arguments
    echo "Error: Invalid argument '$1'." >&2
    echo "Usage: $0 [on|off]" >&2
    exit 1
    ;;
esac

exit 0