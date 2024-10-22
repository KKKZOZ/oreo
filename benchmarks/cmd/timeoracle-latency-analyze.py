import re
import pandas as pd

# Path to your log file
file_path = 'timeoracle.log'

# Read the log file content
with open(file_path, 'r') as f:
    log_content = f.read()

# Regular expression to match LatencyInFunction values
pattern = re.compile(r'"LatencyInFunction":\s"([\d.]+)µs"')

# Extract latency values from log content
latencies = [float(latency) for latency in pattern.findall(log_content)]

# Perform calculations
if latencies:
    max_latency = max(latencies)
    min_latency = min(latencies)
    avg_latency = sum(latencies) / len(latencies)
    count = len(latencies)

    result = {
        "Max Latency (µs)": max_latency,
        "Min Latency (µs)": min_latency,
        "Average Latency (µs)": avg_latency,
        "Total Count": count
    }

    # Convert result to a DataFrame and display
    df_result = pd.DataFrame([result])
    print(df_result)
else:
    print("No latency values found in the log.")
