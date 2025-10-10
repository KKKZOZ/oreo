#!/usr/bin/env python3

import argparse
import os
import subprocess
import sys
import re
import time
from pathlib import Path
import statistics

# Terminal color codes
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
BLUE = "\033[0;34m"
PURPLE = "\033[0;35m"
CYAN = "\033[0;36m"
NC = "\033[0m"

# Global config
executor_port = 8001
timeoracle_port = 8010
thread_load = 100
default_threads = [12, 24, 36, 48, 60, 84, 96]
default_operation_groups = [4,6,8,10,12]
default_record_lengths = [2, 3, 4, 5]
fixed_thread_for_group = 84
fixed_thread_for_len = 84
round_interval = 5
node_list = ["node2", "node3"]
client_nodes = ["node5", "node6"]
password = "kkkzoz"

executor_pid = None
timeoracle_pid = None


def log(msg, color=NC, verbose=True):
    if verbose:
        print(f"{color}{msg}{NC}")


def handle_error(msg):
    print(f"Error: {msg}")
    sys.exit(1)


def kill_process_on_port(port):
    try:
        result = subprocess.run(
            ["lsof", "-t", "-i", f":{port}"], capture_output=True, text=True
        )
        pid = result.stdout.strip()
        if pid:
            print(f"Port {port} is occupied by process {pid}. Terminating...")
            subprocess.run(["kill", "-9", pid])
    except Exception:
        pass


def run_command(cmd, output_file=None, error_file=None):
    stdout = open(output_file, "w") if output_file else subprocess.DEVNULL
    stderr = open(error_file, "w") if error_file else subprocess.DEVNULL

    try:
        result = subprocess.run(cmd, stdout=stdout, stderr=stderr)
        return result.returncode == 0
    finally:
        if output_file:
            stdout.close()
        if error_file:
            stderr.close()


def parse_dsr_values(output_text):
    """Extract all DSR values from output text"""
    dsr_pattern = r"DSR:\s*([0-9.]+)"
    matches = re.findall(dsr_pattern, output_text)
    return [float(val) for val in matches]


def calculate_dsr_statistics(dsr_values):
    """Calculate statistics for DSR values"""
    if not dsr_values:
        return None
    
    sorted_values = sorted(dsr_values)
    p90_index = int(len(sorted_values) * 0.9)
    
    return {
        "count": len(dsr_values),
        "max": max(dsr_values),
        "min": min(dsr_values),
        "avg": statistics.mean(dsr_values),
        "median": statistics.median(dsr_values),
        "p90": sorted_values[p90_index] if sorted_values else 0.0,
        "stdev": statistics.stdev(dsr_values) if len(dsr_values) > 1 else 0.0
    }


def print_dsr_statistics(stats, prefix=""):
    """Print DSR statistics to console"""
    if stats is None:
        log(f"{prefix}No DSR values found", YELLOW, True)
        return
    
    log(f"{prefix}DSR Statistics:", GREEN, True)
    log(f"{prefix}  Count  : {stats['count']}", NC, True)
    log(f"{prefix}  Max    : {stats['max']:.6f}", NC, True)
    log(f"{prefix}  P90    : {stats['p90']:.6f}", NC, True)
    log(f"{prefix}  Median : {stats['median']:.6f}", NC, True)
    log(f"{prefix}  Average: {stats['avg']:.6f}", NC, True)
    log(f"{prefix}  Min    : {stats['min']:.6f}", NC, True)
    log(f"{prefix}  StdDev : {stats['stdev']:.6f}", NC, True)


def save_dsr_statistics(stats, output_file, profile, thread, op_group=None):
    """Save DSR statistics to file"""
    if stats is None:
        return
    
    with open(output_file, "a") as f:
        identifier = f"op{op_group}-" if op_group is not None else ""
        f.write(f"\n{'='*60}\n")
        f.write(f"DSR Statistics for {profile} - {identifier}thread {thread}\n")
        f.write(f"{'='*60}\n")
        f.write(f"Count  : {stats['count']}\n")
        f.write(f"Max    : {stats['max']:.6f}\n")
        f.write(f"P90    : {stats['p90']:.6f}\n")
        f.write(f"Median : {stats['median']:.6f}\n")
        f.write(f"Average: {stats['avg']:.6f}\n")
        f.write(f"Min    : {stats['min']:.6f}\n")
        f.write(f"StdDev : {stats['stdev']:.6f}\n")


def modify_yaml_operation_group(config_file, value):
    """Modify txnoperationgroup value in yaml config file"""
    log(f"Modifying {config_file}: txnoperationgroup = {value}", CYAN, True)
    
    with open(config_file, "r") as f:
        lines = f.readlines()
    
    modified = False
    with open(config_file, "w") as f:
        for line in lines:
            if line.strip().startswith("txnoperationgroup:"):
                f.write(f"txnoperationgroup: {value}\n")
                modified = True
            else:
                f.write(line)
    
    if not modified:
        log(f"Warning: txnoperationgroup not found in {config_file}", YELLOW, True)
    else:
        log(f"Successfully modified txnoperationgroup to {value}", GREEN, True)


def modify_yaml_record_length(config_file, value):
    """Modify max_record_length value in yaml config file"""
    log(f"Modifying {config_file}: max_record_length = {value}", CYAN, True)
    
    with open(config_file, "r") as f:
        lines = f.readlines()
    
    modified = False
    with open(config_file, "w") as f:
        for line in lines:
            if line.strip().startswith("max_record_length:"):
                f.write(f"max_record_length: {value}\n")
                modified = True
            else:
                f.write(line)
    
    if not modified:
        log(f"Warning: max_record_length not found in {config_file}", YELLOW, True)
    else:
        log(f"Successfully modified max_record_length to {value}", GREEN, True)


def parse_specific_errors(output_text):
    """
    Parse and count specific errors:
    - key not found in given RecordLen
    - key not found prev is empty
    """
    error_keywords = [
        "key not found in given RecordLen",
        "key not found prev is empty"
    ]
    
    total_specific_errors = 0
    for keyword in error_keywords:
        count = output_text.count(keyword)
        total_specific_errors += count
    
    return total_specific_errors


def sync_files_to_remote(config, node):
    """Sync necessary files to remote node for multi-client execution"""
    log(f"Syncing files to {node}", CYAN, config["verbose"])
    
    remote_base = "/tmp/oreo-benchmark"
    
    try:
        result = subprocess.run(
            ["ssh", node, f"mkdir -p {remote_base}/bin {remote_base}/workloads/ycsb {remote_base}/config"],
            capture_output=True,
            text=True,
            check=True
        )
        
        log(f"  Syncing binary to {node}...", CYAN, config["verbose"])
        subprocess.run(
            ["rsync", "-az", "./bin/cmd", f"{node}:{remote_base}/bin/"],
            check=True,
            capture_output=True
        )
        
        log(f"  Syncing config files to {node}...", CYAN, config["verbose"])
        subprocess.run(
            ["rsync", "-az", config["config_file"], f"{node}:{remote_base}/workloads/ycsb/"],
            check=True,
            capture_output=True
        )
        
        subprocess.run(
            ["rsync", "-az", config["bc_file"], f"{node}:{remote_base}/config/"],
            check=True,
            capture_output=True
        )
        
        verify_cmd = f"ls -lh {remote_base}/bin/cmd {remote_base}/workloads/ycsb/ {remote_base}/config/"
        result = subprocess.run(
            ["ssh", node, verify_cmd],
            capture_output=True,
            text=True,
            check=True
        )
        
        if config["verbose"]:
            log(f"  Files on {node}:", GREEN, True)
            print(result.stdout)
        
        log(f"Files synced to {node} successfully", GREEN, config["verbose"])
        
    except subprocess.CalledProcessError as e:
        log(f"ERROR: Failed to sync files to {node}", RED, True)
        log(f"Command: {e.cmd}", RED, True)
        log(f"Return code: {e.returncode}", RED, True)
        if e.stdout:
            log(f"stdout: {e.stdout}", RED, True)
        if e.stderr:
            log(f"stderr: {e.stderr}", RED, True)
        raise


def run_workload_remote(node, config, mode, profile, thread, output, is_trace=False):
    """Run workload on a remote node"""
    remote_base = "/tmp/oreo-benchmark"
    
    config_filename = os.path.basename(config["config_file"])
    bc_filename = os.path.basename(config["bc_file"])
    
    cmd = (
        f"cd {remote_base} && ./bin/cmd "
        f"-d oreo-ycsb "
        f"-wl {config['db_combinations']} "
        f"-wc ./workloads/ycsb/{config_filename} "
        f"-bc ./config/{bc_filename} "
        f"-m {mode} "
        f"-ps {profile} "
        f"-t {thread}"
    )
    
    if is_trace:
        cmd += " --trace"
    
    log(f"Running workload on {node}: {profile} thread={thread}", BLUE, config["verbose"])
    
    proc = subprocess.Popen(
        ["ssh", node, cmd],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    return proc


def run_workload(config, mode, profile, thread, output):
    is_trace = (config["wl_mode"] == "trace")
    
    if config.get("multiple_clients", False):
        thread_per_node = thread // 3
        log(
            f"Running {config['wl_type']}-{config['wl_mode']} {profile} "
            f"total_thread={thread} (thread_per_node={thread_per_node})",
            BLUE,
            config["verbose"],
        )
    else:
        thread_per_node = thread
        log(
            f"Running {config['wl_type']}-{config['wl_mode']} {profile} thread={thread}",
            BLUE,
            config["verbose"],
        )

    cmd = [
        "./bin/cmd",
        "-d",
        "oreo-ycsb",
        "-wl",
        config["db_combinations"],
        "-wc",
        config["config_file"],
        "-bc",
        config["bc_file"],
        "-m",
        mode,
        "-ps",
        profile,
        "-t",
        str(thread_per_node),
    ]
    
    if is_trace:
        cmd.append("--trace")

    if config.get("multiple_clients", False):
        processes = []
        
        local_output = output.replace(".txt", "_local.txt")
        log(f"Starting local workload with {thread_per_node} threads", BLUE, config["verbose"])
        
        if is_trace:
            local_proc = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True
            )
        else:
            local_proc = subprocess.Popen(
                cmd,
                stdout=open(local_output, "w"),
                stderr=open(config["log_file"], "a") if config["log_file"] else subprocess.DEVNULL
            )
        processes.append(("local", local_proc, local_output))
        
        for i, node in enumerate(client_nodes):
            node_output = output.replace(".txt", f"_{node}.txt")
            remote_proc = run_workload_remote(node, config, mode, profile, thread_per_node, node_output, is_trace)
            processes.append((node, remote_proc, node_output))
        
        log("Waiting for all client workloads to complete...", YELLOW, config["verbose"])
        all_dsr_values = []
        
        for node_name, proc, node_output in processes:
            stdout, stderr = proc.communicate()
            
            if node_name != "local":
                with open(node_output, "w") as f:
                    f.write(stdout)
                
                error_output = node_output.replace(".txt", "_error.log")
                if stderr:
                    with open(error_output, "w") as f:
                        f.write(stderr)
            
            if is_trace:
                dsr_values = parse_dsr_values(stdout)
                if dsr_values:
                    log(f"Collected {len(dsr_values)} DSR values from {node_name}", CYAN, config["verbose"])
                    all_dsr_values.extend(dsr_values)
            
            if proc.returncode != 0:
                log(f"ERROR: Workload on {node_name} failed with code {proc.returncode}", RED, True)
                if stderr:
                    log(f"Error output from {node_name}:", RED, True)
                    print(stderr)
        
        log("All client workloads completed", GREEN, config["verbose"])
        
        if is_trace and all_dsr_values:
            stats = calculate_dsr_statistics(all_dsr_values)
            print_dsr_statistics(stats, prefix="  ")
            
            dsr_stats_file = output.replace(".txt", "_dsr_stats.txt")
            op_group = config.get("current_op_group")
            save_dsr_statistics(stats, dsr_stats_file, profile, thread, op_group)
            
            return stats
        
    else:
        if is_trace:
            result = subprocess.run(
                cmd,
                stdout=subprocess.PIPE,
                stderr=open(config["log_file"], "a") if config["log_file"] else subprocess.DEVNULL,
                text=True
            )
            
            with open(output, "w") as f:
                f.write(result.stdout)
            
            dsr_values = parse_dsr_values(result.stdout)
            if dsr_values:
                log(f"Collected {len(dsr_values)} DSR values", CYAN, config["verbose"])
                stats = calculate_dsr_statistics(dsr_values)
                print_dsr_statistics(stats, prefix="  ")
                
                dsr_stats_file = output.replace(".txt", "_dsr_stats.txt")
                op_group = config.get("current_op_group")
                save_dsr_statistics(stats, dsr_stats_file, profile, thread, op_group)
                
                return stats
        else:
            run_command(cmd, output, config["log_file"])
    
    return None


def load_data(config):
    for profile in ["native", "cg", "oreo"]:
        log(f"Loading to {config['wl_type']} {profile}", BLUE, config["verbose"])

        cmd = [
            "./bin/cmd",
            "-d",
            "oreo-ycsb",
            "-wl",
            config["db_combinations"],
            "-wc",
            config["config_file"],
            "-bc",
            config["bc_file"],
            "-m",
            "load",
            "-ps",
            profile,
            "-t",
            str(thread_load),
        ]

        subprocess.run(cmd)

    load_flag = Path(config["load_flag_dir"]) / f"{config['wl_type']}-load"
    load_flag.touch()


def get_metrics(config, profile, thread, op_group=None, rec_len=None):
    """
    Get metrics from output file(s).
    For multiple clients, aggregate results from all nodes.
    Returns: (duration, latency, error_ratio, specific_error_count)
    """
    base_name = f"{config['wl_type']}-{config['wl_mode']}-{config['db_combinations']}-{profile}-{thread}"
    if op_group is not None:
        base_name += f"-op{op_group}"
    if rec_len is not None:
        base_name += f"-len{rec_len}"
    
    base_path = f"{config['tar_dir']}/{base_name}"
    
    if config.get("multiple_clients", False):
        node_durations = {}
        node_latencies = {}
        total_success = 0
        total_error = 0
        total_specific_errors = 0
        
        node_names = ["local"] + client_nodes
        for node_name in node_names:
            node_suffix = f"_{node_name}" if node_name != "local" else "_local"
            file_path = f"{base_path}{node_suffix}.txt"
            
            if not os.path.exists(file_path):
                log(f"Warning: Result file not found for {node_name}: {file_path}", RED, True)
                continue
                
            with open(file_path, "r") as f:
                content = f.read()
            
            duration_match = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
            if duration_match:
                node_durations[node_name] = float(duration_match.group(1))
            
            latency_match = re.search(r"99th\(us\): ([0-9]+)", content)
            if latency_match:
                node_latencies[node_name] = int(latency_match.group(1))
            
            success_match = re.search(r"COMMIT .*?Count: ([0-9]+)", content)
            if success_match:
                total_success += int(success_match.group(1))
            
            if profile != "native":
                error_match = re.search(r"COMMIT_ERROR .*?Count: ([0-9]+)", content)
                if error_match:
                    total_error += int(error_match.group(1))
            
            specific_errors = parse_specific_errors(content)
            total_specific_errors += specific_errors
        
        print(f"\n[{profile} thread={thread}] Node metrics:")
        print("  Duration (seconds):")
        for node_name in node_names:
            if node_name in node_durations:
                print(f"    {node_name}: {node_durations[node_name]:.4f}")
        
        print("  Latency p99 (us):")
        for node_name in node_names:
            if node_name in node_latencies:
                print(f"    {node_name}: {node_latencies[node_name]}")
        
        avg_duration = sum(node_durations.values()) / len(node_durations) if node_durations else 0.0
        max_latency = max(node_latencies.values()) if node_latencies else 0
        
        print(f"  Selected average duration: {avg_duration:.4f}")
        print(f"  Selected maximum latency: {max_latency}\n")
        
        total = total_success + total_error
        ratio = total_error / total if total > 0 else 0.0
        
        return avg_duration, max_latency, ratio, total_specific_errors
        
    else:
        file_path = f"{base_path}.txt"
        
        if not os.path.exists(file_path):
            return 0.0, 0, 0.0, 0
        
        with open(file_path, "r") as f:
            content = f.read()

        duration_match = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
        duration = float(duration_match.group(1)) if duration_match else 0.0

        latency_match = re.search(r"99th\(us\): ([0-9]+)", content)
        latency = int(latency_match.group(1)) if latency_match else 0

        success_match = re.search(r"COMMIT .*?Count: ([0-9]+)", content)
        success_cnt = int(success_match.group(1)) if success_match else 0

        error_cnt = 0
        if profile != "native":
            error_match = re.search(r"COMMIT_ERROR .*?Count: ([0-9]+)", content)
            error_cnt = int(error_match.group(1)) if error_match else 0

        total = success_cnt + error_cnt
        ratio = error_cnt / total if total > 0 else 0.0
        
        specific_errors = parse_specific_errors(content)

        return duration, latency, ratio, specific_errors


def print_summary(
    thread,
    native_dur,
    cg_dur,
    oreo_dur,
    native_p99,
    cg_p99,
    oreo_p99,
    native_ratio,
    cg_ratio,
    oreo_ratio,
    op_group=None,
    rec_len=None,
    native_dsr=None,
    cg_dsr=None,
    oreo_dsr=None,
    native_spec_err=0,
    cg_spec_err=0,
    oreo_spec_err=0,
):
    if op_group is not None:
        print(f"Operation Group: {op_group}, Thread: {thread}")
    elif rec_len is not None:
        print(f"Record Length: {rec_len}, Thread: {thread}")
    else:
        print(f"Thread: {thread}")
    
    print(f"native:{native_dur:.4f}")
    print(f"cg    :{cg_dur:.4f}")
    print(f"oreo  :{oreo_dur:.4f}")

    relative_native = oreo_dur / native_dur if native_dur > 0 else 0
    relative_cg = oreo_dur / cg_dur if cg_dur > 0 else 0

    print(f"Oreo:native = {relative_native:.5f}")
    print(f"Oreo:cg     = {relative_cg:.5f}")
    print(f"native 99th: {native_p99}")
    print(f"cg     99th: {cg_p99}")
    print(f"oreo   99th: {oreo_p99}")
    print(f"Error ratio:")
    print(f"native = {native_ratio:.4f}")
    print(f"cg = {cg_ratio:.4f}")
    print(f"oreo = {oreo_ratio:.4f}")
    
    if rec_len is not None:
        print(f"Specific errors (RecordLen related):")
        print(f"native = {native_spec_err}")
        print(f"cg = {cg_spec_err}")
        print(f"oreo = {oreo_spec_err}")
    
    if any([native_dsr, cg_dsr, oreo_dsr]):
        print(f"DSR Statistics:")
        if native_dsr:
            print(f"native - median: {native_dsr['median']:.6f}, p90: {native_dsr['p90']:.6f}, min: {native_dsr['min']:.6f}, max: {native_dsr['max']:.6f}")
        if cg_dsr:
            print(f"cg     - median: {cg_dsr['median']:.6f}, p90: {cg_dsr['p90']:.6f}, min: {cg_dsr['min']:.6f}, max: {cg_dsr['max']:.6f}")
        if oreo_dsr:
            print(f"oreo   - median: {oreo_dsr['median']:.6f}, p90: {oreo_dsr['p90']:.6f}, min: {oreo_dsr['min']:.6f}, max: {oreo_dsr['max']:.6f}")
    
    print("---------------------------------")


def deploy_local(config):
    global executor_pid, timeoracle_pid

    kill_process_on_port(executor_port)
    kill_process_on_port(timeoracle_port)

    log("Starting executor", NC, config["verbose"])
    executor_log = open("./log/executor.log", "w")
    executor_proc = subprocess.Popen(
        [
            "./bin/executor",
            "-p",
            str(executor_port),
            "-w",
            config["wl_type"],
            "-bc",
            config["bc_file"],
            "-db",
            config["db_combinations"],
        ],
        stderr=executor_log,
    )
    executor_pid = executor_proc.pid

    log("Starting time oracle", NC, config["verbose"])
    timeoracle_log = open("./log/timeoracle.log", "w")
    timeoracle_proc = subprocess.Popen(
        ["./bin/timeoracle", "-p", str(timeoracle_port), "-type", "hybrid"],
        stdout=subprocess.DEVNULL,
        stderr=timeoracle_log,
    )
    timeoracle_pid = timeoracle_proc.pid


def deploy_remote(config):
    log("Setup timeoracle on node 2", GREEN, config["verbose"])
    
    if config["wl_mode"] == "trace":
        timeoracle_type = "simple"
        log("Using 'simple' timeoracle type for trace mode", YELLOW, config["verbose"])
    else:
        timeoracle_type = ""
    
    cmd = f"echo '{password}' | sudo -S bash /root/oreo-ben/start-timeoracle.sh {timeoracle_type}".strip()
    subprocess.run(["ssh", "-t", node_list[0], cmd])

    for node in node_list:
        log(f"Setup {node}", GREEN, config["verbose"])
        cmd = (
            f"echo '{password}' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh "
            f"-p 8001 -wl {config['wl_type']} -db {config['db_combinations']} "
            f"-bc BenConfig_ycsb.yaml -r"
        )
        subprocess.run(["ssh", "-t", node, cmd])


def cleanup(config):
    if not config["remote"]:
        if executor_pid:
            log("Killing executor", NC, config["verbose"])
            subprocess.run(["kill", str(executor_pid)])
        if timeoracle_pid:
            log("Killing time oracle", NC, config["verbose"])
            subprocess.run(["kill", str(timeoracle_pid)])


def extract_operation_count(config_file):
    with open(config_file, "r") as f:
        for line in f:
            match = re.match(r"^operationcount.*?([0-9.]+)", line)
            if match:
                return match.group(1)
    return "0"


def parse_args():
    parser = argparse.ArgumentParser(description="Benchmark script")
    parser.add_argument("-wl", "--workload", required=True, help="Workload mode")
    parser.add_argument("-db", "--db", required=True, help="Database combinations")
    parser.add_argument("-t", "--threads", help="Thread counts (comma-separated)")
    parser.add_argument(
        "-og", "--operation-groups", 
        help="Operation group counts for 'group' workload mode (comma-separated, default: 4,6,8,10,12)"
    )
    parser.add_argument(
        "-rl", "--record-lengths",
        help="Record lengths for 'len' workload mode (comma-separated, default: 2,3,4,5)"
    )
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    parser.add_argument("-r", "--remote", action="store_true", help="Remote execution")
    parser.add_argument("-s", "--skip", action="store_true", help="Skip deployment")
    parser.add_argument(
        "-l", "--loaded", action="store_true", help="Data already loaded"
    )
    parser.add_argument(
        "-mc", "--multiple-clients", action="store_true", 
        help="Run workload on multiple client nodes (node5, node6) simultaneously"
    )

    return parser.parse_args()


def main():
    args = parse_args()

    os.chdir(Path(__file__).parent.parent)

    is_group_mode = (args.workload == "group")
    is_trace_mode = (args.workload == "trace")
    is_len_mode = (args.workload == "len")
    
    if is_group_mode:
        threads = [fixed_thread_for_group]
        operation_groups = (
            [int(og.strip()) for og in args.operation_groups.split(",")]
            if args.operation_groups
            else default_operation_groups
        )
        record_lengths = None
        log(f"Group mode detected: using fixed thread count {fixed_thread_for_group}", YELLOW, True)
        log(f"Operation groups to test: {operation_groups}", YELLOW, True)
    elif is_len_mode:
        threads = [fixed_thread_for_len]
        operation_groups = None
        record_lengths = (
            [int(rl.strip()) for rl in args.record_lengths.split(",")]
            if args.record_lengths
            else default_record_lengths
        )
        log(f"Len mode detected: using fixed thread count {fixed_thread_for_len}", YELLOW, True)
        log(f"Record lengths to test: {record_lengths}", YELLOW, True)
    else:
        threads = (
            [int(t.strip()) for t in args.threads.split(",")]
            if args.threads
            else default_threads
        )
        operation_groups = None
        record_lengths = None
    
    if is_trace_mode:
        log("Trace mode detected: will collect DSR statistics", YELLOW, True)
    
    if args.multiple_clients:
        invalid_threads = [t for t in threads if t % 3 != 0]
        if invalid_threads:
            handle_error(
                f"In multiple-clients mode, all thread counts must be divisible by 3. "
                f"Invalid values: {invalid_threads}"
            )

    wl_type = "ycsb"
    load_flag_dir = "./data/ycsb-micro"
    tar_dir = "./data/ycsb-micro"

    config = {
        "wl_type": wl_type,
        "wl_mode": args.workload,
        "db_combinations": args.db,
        "threads": threads,
        "verbose": args.verbose,
        "remote": args.remote,
        "skip": args.skip,
        "loaded": args.loaded,
        "multiple_clients": args.multiple_clients,
        "load_flag_dir": load_flag_dir,
        "tar_dir": f"{tar_dir}/{args.workload}-{args.db}",
        "config_file": f"./workloads/ycsb/{args.workload}_{args.db}.yaml",
        "bc_file": "./config/BenConfig_ycsb.yaml",
        "log_file": None,
    }

    if not Path(config["config_file"]).exists():
        handle_error(f"Config file {config['config_file']} does not exist")
    if not Path(config["bc_file"]).exists():
        handle_error(f"Config file {config['bc_file']} does not exist")

    print("Building the benchmark executable")
    subprocess.run(["go", "build", "."])
    Path("cmd").rename("./bin/cmd")

    Path(config["tar_dir"]).mkdir(parents=True, exist_ok=True)
    config["log_file"] = f"{config['tar_dir']}/benchmark.log"

    mode_suffix = "_multi" if config["multiple_clients"] else ""
    group_suffix = "_group" if is_group_mode else ""
    len_suffix = "_len" if is_len_mode else ""
    trace_suffix = "_trace" if is_trace_mode else ""
    results_file = (
        f"{config['tar_dir']}/{args.workload}_{args.db}_benchmark_results"
        f"{mode_suffix}{group_suffix}{len_suffix}{trace_suffix}.csv"
    )
    
    with open(results_file, "w") as f:
        if is_group_mode:
            header = "operation_group,thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err"
            if is_trace_mode:
                header += ",native_dsr_median,native_dsr_p90,native_dsr_min,native_dsr_max,cg_dsr_median,cg_dsr_p90,cg_dsr_min,cg_dsr_max,oreo_dsr_median,oreo_dsr_p90,oreo_dsr_min,oreo_dsr_max"
            f.write(header + "\n")
        elif is_len_mode:
            header = "record_length,thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err,native_spec_err,cg_spec_err,oreo_spec_err"
            if is_trace_mode:
                header += ",native_dsr_median,native_dsr_p90,native_dsr_min,native_dsr_max,cg_dsr_median,cg_dsr_p90,cg_dsr_min,cg_dsr_max,oreo_dsr_median,oreo_dsr_p90,oreo_dsr_min,oreo_dsr_max"
            f.write(header + "\n")
        else:
            header = "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err"
            if is_trace_mode:
                header += ",native_dsr_median,native_dsr_p90,native_dsr_min,native_dsr_max,cg_dsr_median,cg_dsr_p90,cg_dsr_min,cg_dsr_max,oreo_dsr_median,oreo_dsr_p90,oreo_dsr_min,oreo_dsr_max"
            f.write(header + "\n")

    operation = extract_operation_count(config["config_file"])

    log(
        f"Running benchmark for [{wl_type}] workload with [{args.db}] database combinations",
        YELLOW,
        config["verbose"],
    )
    
    if config["multiple_clients"]:
        log("Multiple client mode enabled - will run on local, node5, and node6", YELLOW, True)
        log(f"Thread counts will be divided by 3 across nodes: {threads}", YELLOW, True)

    if config["skip"]:
        log("Skipping deployment", YELLOW, config["verbose"])
    else:
        if config["remote"]:
            print("Running remotely")
            deploy_remote(config)
        else:
            print("Running locally")
            deploy_local(config)

    if config["multiple_clients"]:
        log("Syncing files to client nodes...", YELLOW, True)
        
        for node in client_nodes:
            log(f"Testing SSH connection to {node}...", CYAN, True)
            try:
                result = subprocess.run(
                    ["ssh", "-o", "ConnectTimeout=5", node, "echo 'Connection OK'"],
                    capture_output=True,
                    text=True,
                    check=True,
                    timeout=10
                )
                log(f"  Connection to {node}: OK", GREEN, True)
            except subprocess.TimeoutExpired:
                handle_error(f"SSH connection to {node} timed out")
            except subprocess.CalledProcessError as e:
                handle_error(f"Cannot connect to {node}: {e.stderr}")
        
        for node in client_nodes:
            sync_files_to_remote(config, node)

    load_flag_file = Path(load_flag_dir) / f"{wl_type}-load"
    if config["loaded"]:
        log("Skipping data loading", YELLOW, config["verbose"])
    else:
        if load_flag_file.exists():
            log("Data has been already loaded", YELLOW, config["verbose"])
        else:
            log("Ready to load data", YELLOW, config["verbose"])

    if not load_flag_file.exists():
        load_data(config)

    profiles = ["native", "cg", "oreo"]
    if is_trace_mode:
        profiles = ["oreo"]
        log("Trace mode: only running 'oreo' profile", YELLOW, True)

    if is_group_mode:
        for op_group in operation_groups:
            log(f"\n{'='*60}", PURPLE, True)
            log(f"Testing with operation group: {op_group}", PURPLE, True)
            log(f"{'='*60}\n", PURPLE, True)
            
            modify_yaml_operation_group(config["config_file"], op_group)
            config["current_op_group"] = op_group
            
            if config["multiple_clients"]:
                for node in client_nodes:
                    log(f"Syncing updated config to {node}...", CYAN, config["verbose"])
                    sync_files_to_remote(config, node)
            
            thread = fixed_thread_for_group
            dsr_stats = {"native": None, "cg": None, "oreo": None}
            
            for profile in profiles:
                output = (
                    f"{config['tar_dir']}/{wl_type}-{args.workload}-"
                    f"{args.db}-{profile}-{thread}-op{op_group}.txt"
                )
                dsr_stat = run_workload(config, "run", profile, thread, output)
                if dsr_stat:
                    dsr_stats[profile] = dsr_stat

            native_dur, native_p99, native_ratio, _ = get_metrics(config, "native", thread, op_group)
            cg_dur, cg_p99, cg_ratio, _ = get_metrics(config, "cg", thread, op_group)
            oreo_dur, oreo_p99, oreo_ratio, _ = get_metrics(config, "oreo", thread, op_group)

            with open(results_file, "a") as f:
                line = (
                    f"{op_group},{thread},{operation},{native_dur:.4f},{cg_dur:.4f},{oreo_dur:.4f},"
                    f"{native_p99},{cg_p99},{oreo_p99},"
                    f"{native_ratio:.4f},{cg_ratio:.4f},{oreo_ratio:.4f}"
                )
                
                if is_trace_mode:
                    for profile in ["native", "cg", "oreo"]:
                        if dsr_stats[profile]:
                            s = dsr_stats[profile]
                            line += f",{s['median']:.6f},{s['p90']:.6f},{s['min']:.6f},{s['max']:.6f}"
                        else:
                            line += ",,,,"
                
                f.write(line + "\n")

            print_summary(
                thread,
                native_dur,
                cg_dur,
                oreo_dur,
                native_p99,
                cg_p99,
                oreo_p99,
                native_ratio,
                cg_ratio,
                oreo_ratio,
                op_group=op_group,
                native_dsr=dsr_stats["native"],
                cg_dsr=dsr_stats["cg"],
                oreo_dsr=dsr_stats["oreo"],
            )

            time.sleep(round_interval)
            
    elif is_len_mode:
        for rec_len in record_lengths:
            log(f"\n{'='*60}", PURPLE, True)
            log(f"Testing with record length: {rec_len}", PURPLE, True)
            log(f"{'='*60}\n", PURPLE, True)
            
            modify_yaml_record_length(config["config_file"], rec_len)
            config["current_rec_len"] = rec_len
            
            if config["multiple_clients"]:
                for node in client_nodes:
                    log(f"Syncing updated config to {node}...", CYAN, config["verbose"])
                    sync_files_to_remote(config, node)
            
            thread = fixed_thread_for_len
            dsr_stats = {"native": None, "cg": None, "oreo": None}
            
            for profile in profiles:
                output = (
                    f"{config['tar_dir']}/{wl_type}-{args.workload}-"
                    f"{args.db}-{profile}-{thread}-len{rec_len}.txt"
                )
                dsr_stat = run_workload(config, "run", profile, thread, output)
                if dsr_stat:
                    dsr_stats[profile] = dsr_stat

            native_dur, native_p99, native_ratio, native_spec_err = get_metrics(
                config, "native", thread, rec_len=rec_len
            )
            cg_dur, cg_p99, cg_ratio, cg_spec_err = get_metrics(
                config, "cg", thread, rec_len=rec_len
            )
            oreo_dur, oreo_p99, oreo_ratio, oreo_spec_err = get_metrics(
                config, "oreo", thread, rec_len=rec_len
            )

            with open(results_file, "a") as f:
                line = (
                    f"{rec_len},{thread},{operation},{native_dur:.4f},{cg_dur:.4f},{oreo_dur:.4f},"
                    f"{native_p99},{cg_p99},{oreo_p99},"
                    f"{native_ratio:.4f},{cg_ratio:.4f},{oreo_ratio:.4f},"
                    f"{native_spec_err},{cg_spec_err},{oreo_spec_err}"
                )
                
                if is_trace_mode:
                    for profile in ["native", "cg", "oreo"]:
                        if dsr_stats[profile]:
                            s = dsr_stats[profile]
                            line += f",{s['median']:.6f},{s['p90']:.6f},{s['min']:.6f},{s['max']:.6f}"
                        else:
                            line += ",,,,"
                
                f.write(line + "\n")

            print_summary(
                thread,
                native_dur,
                cg_dur,
                oreo_dur,
                native_p99,
                cg_p99,
                oreo_p99,
                native_ratio,
                cg_ratio,
                oreo_ratio,
                rec_len=rec_len,
                native_dsr=dsr_stats["native"],
                cg_dsr=dsr_stats["cg"],
                oreo_dsr=dsr_stats["oreo"],
                native_spec_err=native_spec_err,
                cg_spec_err=cg_spec_err,
                oreo_spec_err=oreo_spec_err,
            )

            time.sleep(round_interval)
    else:
        for thread in threads:
            dsr_stats = {"native": None, "cg": None, "oreo": None}
            
            for profile in profiles:
                output = (
                    f"{config['tar_dir']}/{wl_type}-{args.workload}-"
                    f"{args.db}-{profile}-{thread}.txt"
                )
                dsr_stat = run_workload(config, "run", profile, thread, output)
                if dsr_stat:
                    dsr_stats[profile] = dsr_stat

            native_dur, native_p99, native_ratio, _ = get_metrics(config, "native", thread)
            cg_dur, cg_p99, cg_ratio, _ = get_metrics(config, "cg", thread)
            oreo_dur, oreo_p99, oreo_ratio, _ = get_metrics(config, "oreo", thread)

            with open(results_file, "a") as f:
                line = (
                    f"{thread},{operation},{native_dur:.4f},{cg_dur:.4f},{oreo_dur:.4f},"
                    f"{native_p99},{cg_p99},{oreo_p99},"
                    f"{native_ratio:.4f},{cg_ratio:.4f},{oreo_ratio:.4f}"
                )
                
                if is_trace_mode:
                    for profile in ["native", "cg", "oreo"]:
                        if dsr_stats[profile]:
                            s = dsr_stats[profile]
                            line += f",{s['median']:.6f},{s['p90']:.6f},{s['min']:.6f},{s['max']:.6f}"
                        else:
                            line += ",,,,"
                
                f.write(line + "\n")

            print_summary(
                thread,
                native_dur,
                cg_dur,
                oreo_dur,
                native_p99,
                cg_p99,
                oreo_p99,
                native_ratio,
                cg_ratio,
                oreo_ratio,
                native_dsr=dsr_stats["native"],
                cg_dsr=dsr_stats["cg"],
                oreo_dsr=dsr_stats["oreo"],
            )

            time.sleep(round_interval)

    cleanup(config)
    
    log(f"\nBenchmark completed! Results saved to: {results_file}", GREEN, True)


if __name__ == "__main__":
    main()