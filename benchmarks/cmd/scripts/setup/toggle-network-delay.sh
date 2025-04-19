#!/bin/bash

# Script to manage bidirectional network delay using tc and IFB
# Usage: sudo ./toggle-netwoek-delay.sh [on|off]

# --- Configuration ---
IFACE="eth0"         # Primary network interface (CHANGE IF NEEDED)
IFB_IFACE="ifb0"     # IFB interface to use
DELAY="1.5ms"        # One-way delay to add (results in 2x DELAY RTT increase)
# --- End Configuration ---

# Check for root privileges
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root. Use sudo."
   exit 1
fi

# Check for correct number of arguments
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 [on|off]"
    exit 1
fi

ACTION="$1"

# Function to apply the delay rules
apply_delay() {
    echo "Applying ${DELAY} bidirectional delay on ${IFACE} via ${IFB_IFACE}..."

    # 1. Load ifb module
    echo "Loading ifb module..."
    modprobe ifb numifbs=1
    if [ $? -ne 0 ]; then echo "Error loading ifb module. Aborting."; exit 1; fi

    # 2. Bring IFB interface up
    echo "Bringing up ${IFB_IFACE}..."
    ip link set dev ${IFB_IFACE} up
    if [ $? -ne 0 ]; then echo "Error bringing up ${IFB_IFACE}. Aborting."; exit 1; fi

    # 3. Add ingress qdisc to physical interface (needed for filter)
    echo "Adding ingress qdisc to ${IFACE}..."
    tc qdisc add dev ${IFACE} handle ffff: ingress
    # Ignore errors if it already exists, proceed cautiously

    # 4. Redirect ingress traffic from physical interface to IFB
    echo "Redirecting ingress traffic from ${IFACE} to ${IFB_IFACE}..."
    # Delete potential stale filter first, then add. Best effort.
    tc filter del dev ${IFACE} parent ffff: > /dev/null 2>&1
    tc filter add dev ${IFACE} parent ffff: protocol all u32 \
        match u32 0 0 \
        action mirred egress redirect dev ${IFB_IFACE}
    if [ $? -ne 0 ]; then echo "Error adding redirection filter. Aborting."; exit 1; fi

    # 5. Add netem delay qdisc to IFB interface (for ingress traffic)
    # Use 'replace' to handle existing qdisc or create new one
    echo "Applying ${DELAY} delay to ingress traffic via ${IFB_IFACE}..."
    tc qdisc replace dev ${IFB_IFACE} root netem delay ${DELAY}
    if [ $? -ne 0 ]; then echo "Error adding netem qdisc to ${IFB_IFACE}. Aborting."; exit 1; fi

    # 6. Add netem delay qdisc to physical interface (for egress traffic)
    # Use 'replace' to handle existing qdisc or create new one
    echo "Applying ${DELAY} delay to egress traffic on ${IFACE}..."
    tc qdisc replace dev ${IFACE} root netem delay ${DELAY}
    if [ $? -ne 0 ]; then echo "Error adding netem qdisc to ${IFACE}. Aborting."; exit 1; fi

    echo "Bidirectional delay rules applied successfully."
}

# Function to remove the delay rules
remove_delay() {
    echo "Removing bidirectional delay rules from ${IFACE} and ${IFB_IFACE}..."

    # Important: Remove in reverse order of complexity / dependencies

    # 1. Remove egress qdisc from physical interface
    echo "Removing egress qdisc from ${IFACE}..."
    tc qdisc del dev ${IFACE} root > /dev/null 2>&1
    # Ignore errors if not found

    # 2. Remove egress qdisc from IFB interface
    echo "Removing egress qdisc from ${IFB_IFACE}..."
    tc qdisc del dev ${IFB_IFACE} root > /dev/null 2>&1
    # Ignore errors if not found

    # 3. Remove redirection filter from physical interface ingress
    echo "Removing redirection filter from ${IFACE} ingress..."
    tc filter del dev ${IFACE} parent ffff: > /dev/null 2>&1
    # Ignore errors if not found

    # 4. Remove ingress qdisc from physical interface
    echo "Removing ingress qdisc from ${IFACE}..."
    tc qdisc del dev ${IFACE} handle ffff: ingress > /dev/null 2>&1
    # Ignore errors if not found

    # 5. Bring IFB interface down (optional, but good practice)
    echo "Bringing down ${IFB_IFACE} (optional)..."
    ip link set dev ${IFB_IFACE} down > /dev/null 2>&1

    # 6. Unload ifb module (optional, if no longer needed by anything else)
    # echo "Attempting to unload ifb module (optional)..."
    # modprobe -r ifb > /dev/null 2>&1

    echo "Delay rules removed."
}

# Main logic based on argument
case "$ACTION" in
    on)
        apply_delay
        ;;
    off)
        remove_delay
        ;;
    *)
        echo "Invalid argument: $ACTION"
        echo "Usage: $0 [on|off]"
        exit 1
        ;;
esac

exit 0