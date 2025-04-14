import csv
import argparse
import os
from collections import defaultdict
import math
import sys # For sys.maxsize

def find_min_start_time(input_filename):
    """
    Performs the first pass to find the minimum timestamp among
    all TXN_OK and TXN_ERROR operations.

    Args:
        input_filename (str): Path to the input CSV file.

    Returns:
        int or None: The minimum timestamp in microseconds, or None if
                     no valid transactions are found or an error occurs.
    """
    min_ts_us = None
    line_num = 0
    try:
        with open(input_filename, 'r', newline='') as infile:
            reader = csv.reader(infile)
            try:
                header = next(reader) # Read header row
                line_num += 1
                # Optional: header validation
                expected_header = ['operation', 'timestamp_us', 'latency_us']
                if header != expected_header:
                    print(f"Warning (Pass 1): Unexpected header format. Expected {expected_header}, got {header}.")
            except StopIteration:
                print("Error (Pass 1): Input file is empty.")
                return None

            for row in reader:
                line_num += 1
                if len(row) != 3:
                    # Silently skip malformed rows in this pass or add warning
                    # print(f"Warning (Pass 1): Skipping malformed row {line_num}: {row}")
                    continue

                operation = row[0]

                # Only consider relevant transaction types
                if operation in ('TXN_OK', 'TXN_ERROR'):
                    try:
                        timestamp_us = int(row[1])
                        if min_ts_us is None or timestamp_us < min_ts_us:
                            min_ts_us = timestamp_us
                    except ValueError:
                        # Silently skip rows with bad timestamps in this pass or add warning
                        # print(f"Warning (Pass 1): Skipping row {line_num} with non-integer timestamp: {row}")
                        continue
                    except IndexError:
                         # Should be caught by len(row) check, but defensive
                         # print(f"Warning (Pass 1): Skipping row {line_num} with incorrect columns: {row}")
                         continue

    except FileNotFoundError:
        print(f"Error (Pass 1): Input file '{input_filename}' not found.")
        return None
    except Exception as e:
        print(f"An unexpected error occurred during Pass 1 (finding min timestamp): {e}")
        return None

    if min_ts_us is None:
        print("Warning (Pass 1): No 'TXN_OK' or 'TXN_ERROR' operations found in the file.")

    return min_ts_us


def process_csv(input_filename, span_ms):
    """
    Processes a CSV file containing transaction logs, aggregates counts
    by time span based on the overall minimum transaction timestamp,
    and writes the results to a new CSV file. Uses a two-pass approach.

    Args:
        input_filename (str): Path to the input CSV file.
        span_ms (int): Time interval span in milliseconds.
    """
    if not os.path.exists(input_filename):
        # This check is slightly redundant with find_min_start_time, but good practice
        print(f"Error: Input file '{input_filename}' not found.")
        return

    if span_ms <= 0:
        print("Error: Span must be a positive integer.")
        return

    # --- Pass 1: Find the minimum start time ---
    print(f"Starting Pass 1: Finding minimum transaction timestamp in '{input_filename}'...")
    start_time_us = find_min_start_time(input_filename)

    if start_time_us is None:
        print("Cannot proceed without a valid start time. Exiting.")
        return
    print(f"Pass 1 Complete: Minimum transaction timestamp found: {start_time_us} us.")

    # --- Pass 2: Aggregate counts based on the determined start time ---
    print("Starting Pass 2: Aggregating counts...")

    span_us = span_ms * 1000
    interval_counts = defaultdict(lambda: {'ok': 0, 'error': 0})
    max_interval_index = -1
    processed_rows = 0

    try:
        with open(input_filename, 'r', newline='') as infile:
            reader = csv.reader(infile)
            try:
                header = next(reader) # Skip header row again
            except StopIteration:
                # Should have been caught in Pass 1, but defensive check
                print("Error (Pass 2): Input file appears empty (only header?).")
                return

            line_num = 1 # Start after header
            for row in reader:
                line_num += 1
                if len(row) != 3:
                    print(f"Warning (Pass 2): Skipping malformed row {line_num}: {row}")
                    continue

                operation = row[0]

                # Only process relevant transaction types
                if operation in ('TXN_OK', 'TXN_ERROR'):
                    try:
                        timestamp_us = int(row[1])

                        # Calculate time elapsed since the *global minimum* start time
                        relative_time_us = timestamp_us - start_time_us

                        # Timestamps should not be before the determined start time
                        if relative_time_us < 0:
                             print(f"Warning (Pass 2): Row {line_num}: Timestamp {timestamp_us} is earlier than overall minimum start time {start_time_us}. This is unexpected. Skipping row.")
                             continue

                        # Determine the interval index
                        interval_index = relative_time_us // span_us

                        # Update the maximum interval index
                        max_interval_index = max(max_interval_index, interval_index)

                        # Increment the count
                        if operation == 'TXN_OK':
                            interval_counts[interval_index]['ok'] += 1
                        else: # TXN_ERROR
                            interval_counts[interval_index]['error'] += 1
                        processed_rows += 1

                    except ValueError:
                        print(f"Warning (Pass 2): Skipping row {line_num} with non-integer timestamp: {row}")
                    except IndexError:
                         print(f"Warning (Pass 2): Skipping row {line_num} with incorrect columns: {row}")
                # else: # Optionally skip other operation types silently or log them
                    # print(f"Debug (Pass 2): Skipping non-transaction operation '{operation}' at line {line_num}.")


    except FileNotFoundError:
        # Should not happen if Pass 1 succeeded, but defensive
        print(f"Error (Pass 2): Input file '{input_filename}' seems to have disappeared.")
        return
    except Exception as e:
        print(f"An unexpected error occurred during Pass 2 (aggregating counts): {e}")
        return

    print(f"Pass 2 Complete: Processed {processed_rows} transaction rows.")

    # --- Prepare and write the output ---
    if max_interval_index == -1:
         # This case means start_time_us was found, but maybe all rows had issues in Pass 2
         # or relative_time_us was always negative (which would be strange).
         print("No data points were successfully aggregated in Pass 2. Output file will not be generated.")
         return


    output_filename = os.path.splitext(input_filename)[0] + f"_processed_span_{span_ms}ms.csv"

    try:
        with open(output_filename, 'w', newline='') as outfile:
            writer = csv.writer(outfile)
            writer.writerow(['from_start', 'ok', 'error']) # Write header

            # Iterate through all intervals from 0 up to the maximum index found
            for i in range(max_interval_index + 1):
                from_start_ms = (i + 1) * span_ms
                counts = interval_counts[i] # Uses defaultdict's default if key not found
                ok_count = counts['ok']
                error_count = counts['error']
                writer.writerow([from_start_ms, ok_count, error_count])

        print(f"Processing complete. Output written to '{output_filename}'")

    except Exception as e:
        print(f"An error occurred while writing the output file: {e}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Process transaction log CSV to count operations per time span, using the earliest transaction time as the start.")
    parser.add_argument("-f", "--file", required=True, help="Path to the input CSV file.")
    parser.add_argument("-s", "--span", required=True, type=int, help="Time span in milliseconds (ms).")

    args = parser.parse_args()

    process_csv(args.file, args.span)