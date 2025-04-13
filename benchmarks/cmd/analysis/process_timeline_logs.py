import csv
import argparse
import os
from collections import defaultdict
import math

def process_csv(input_filename, span_ms):
    """
    Processes a CSV file containing transaction logs, aggregates counts
    by time span, and writes the results to a new CSV file.

    Args:
        input_filename (str): Path to the input CSV file.
        span_ms (int): Time interval span in milliseconds.
    """
    if not os.path.exists(input_filename):
        print(f"Error: Input file '{input_filename}' not found.")
        return

    if span_ms <= 0:
        print("Error: Span must be a positive integer.")
        return

    # Convert span from milliseconds to microseconds for calculations
    span_us = span_ms * 1000

    # Use defaultdict to easily store counts per interval
    # Key: interval index (0, 1, 2, ...)
    # Value: dictionary {'ok': count, 'error': count}
    interval_counts = defaultdict(lambda: {'ok': 0, 'error': 0})

    start_time_us = None
    max_interval_index = -1

    try:
        with open(input_filename, 'r', newline='') as infile:
            reader = csv.reader(infile)
            header = next(reader) # Skip header row

            # Verify header format (optional but good practice)
            expected_header = ['operation', 'timestamp_us', 'latency_us']
            if header != expected_header:
                 print(f"Warning: Unexpected header format. Expected {expected_header}, got {header}. Proceeding anyway.")


            first_row = True
            for row in reader:
                if len(row) != 3:
                    print(f"Warning: Skipping malformed row: {row}")
                    continue

                try:
                    operation = row[0]
                    timestamp_us = int(row[1])
                    # latency_us = int(row[2]) # Latency is not used in this calculation

                    # Set the start time based on the first valid data row
                    if first_row:
                        start_time_us = timestamp_us
                        first_row = False

                    # Ensure start_time_us is set (handles empty files or files with only headers)
                    if start_time_us is None:
                         print("Error: No valid data rows found after header.")
                         return

                    # Calculate time elapsed since the start in microseconds
                    relative_time_us = timestamp_us - start_time_us

                    # Determine the interval index this timestamp falls into
                    # Example: span_us = 100000
                    # relative_time_us = 50000  -> index = 0 (Interval ends at 1 * span_us)
                    # relative_time_us = 150000 -> index = 1 (Interval ends at 2 * span_us)
                    # relative_time_us = 99999  -> index = 0
                    # relative_time_us = 100000 -> index = 1
                    # Use floor division //
                    if relative_time_us < 0:
                         # This shouldn't happen if timestamps are increasing, but handle defensively
                         print(f"Warning: Timestamp {timestamp_us} is earlier than start time {start_time_us}. Skipping.")
                         continue

                    # The interval index is based on *when the interval starts* relative to the time elapsed
                    interval_index = relative_time_us // span_us

                    # Update the maximum interval index seen so far
                    max_interval_index = max(max_interval_index, interval_index)

                    # Increment the count for the correct operation type in the calculated interval
                    if operation == 'TXN_OK':
                        interval_counts[interval_index]['ok'] += 1
                    elif operation == 'TXN_ERROR':
                        interval_counts[interval_index]['error'] += 1
                    else:
                        print(f"Warning: Skipping unknown operation type '{operation}' in row: {row}")

                except ValueError:
                    print(f"Warning: Skipping row with non-integer timestamp/latency: {row}")
                except IndexError:
                     print(f"Warning: Skipping row with incorrect number of columns: {row}")


    except FileNotFoundError:
        print(f"Error: Input file '{input_filename}' not found.")
        return
    except StopIteration:
        print("Error: Input file is empty or contains only a header.")
        return
    except Exception as e:
        print(f"An unexpected error occurred while reading the file: {e}")
        return

    # --- Prepare and write the output ---
    if start_time_us is None:
        print("No data processed. Output file will not be generated.")
        return

    output_filename = os.path.splitext(input_filename)[0] + f"_processed_span_{span_ms}ms.csv"

    try:
        with open(output_filename, 'w', newline='') as outfile:
            writer = csv.writer(outfile)
            writer.writerow(['from_start', 'ok', 'error']) # Write header

            # Iterate through all intervals from 0 up to the maximum index found
            # This ensures even empty intervals are included in the output
            for i in range(max_interval_index + 1):
                # Calculate the end time of the interval in milliseconds relative to start
                from_start_ms = (i + 1) * span_ms

                # Get the counts for this interval (defaults to {'ok': 0, 'error': 0} if not present)
                counts = interval_counts[i]
                ok_count = counts['ok']
                error_count = counts['error']

                writer.writerow([from_start_ms, ok_count, error_count])

        print(f"Processing complete. Output written to '{output_filename}'")

    except Exception as e:
        print(f"An error occurred while writing the output file: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Process transaction log CSV to count operations per time span.")
    parser.add_argument("-f", "--file", required=True, help="Path to the input CSV file.")
    parser.add_argument("-s", "--span", required=True, type=int, help="Time span in milliseconds (ms).")

    args = parser.parse_args()

    process_csv(args.file, args.span)