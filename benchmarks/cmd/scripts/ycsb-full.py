#!/usr/bin/env python3

import argparse
import os
import subprocess
import sys
import re
import time
from pathlib import Path

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


def sync_files_to_remote(config, node):
    """Sync necessary files to remote node for multi-client execution"""
    log(f"Syncing files to {node}", CYAN, config["verbose"])
    
    remote_base = "/tmp/oreo-benchmark"
    
    try:
        # Create remote directories
        result = subprocess.run(
            ["ssh", node, f"mkdir -p {remote_base}/bin {remote_base}/workloads/ycsb {remote_base}/config"],
            capture_output=True,
            text=True,
            check=True
        )
        
        # Sync binary file
        log(f"  Syncing binary to {node}...", CYAN, config["verbose"])
        subprocess.run(
            ["rsync", "-az", "./bin/cmd", f"{node}:{remote_base}/bin/"],
            check=True,
            capture_output=True
        )
        
        # Sync config files
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
        
        # Verify files exist on remote
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


def run_workload_remote(node, config, mode, profile, thread, output):
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
    
    log(f"Running workload on {node}: {profile} thread={thread}", BLUE, config["verbose"])
    
    # Execute command on remote node and redirect output
    proc = subprocess.Popen(
        ["ssh", node, cmd],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    return proc


def run_workload(config, mode, profile, thread, output):
    if config.get("multiple_clients", False):
        # Divide threads equally among 3 nodes
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

    if config.get("multiple_clients", False):
        # Run on multiple nodes in parallel
        processes = []
        
        # Run on local node
        local_output = output.replace(".txt", "_local.txt")
        log(f"Starting local workload with {thread_per_node} threads", BLUE, config["verbose"])
        local_proc = subprocess.Popen(
            cmd,
            stdout=open(local_output, "w"),
            stderr=open(config["log_file"], "a") if config["log_file"] else subprocess.DEVNULL
        )
        processes.append(("local", local_proc, local_output))
        
        # Run on remote client nodes
        for i, node in enumerate(client_nodes):
            node_output = output.replace(".txt", f"_{node}.txt")
            remote_proc = run_workload_remote(node, config, mode, profile, thread_per_node, node_output)
            processes.append((node, remote_proc, node_output))
        
        # Wait for all processes to complete
        log("Waiting for all client workloads to complete...", YELLOW, config["verbose"])
        for node_name, proc, node_output in processes:
            stdout, stderr = proc.communicate()
            
            # Write remote output to file
            if node_name != "local":
                with open(node_output, "w") as f:
                    f.write(stdout)
                
                # Save stderr to error log
                error_output = node_output.replace(".txt", "_error.log")
                if stderr:
                    with open(error_output, "w") as f:
                        f.write(stderr)
            
            if proc.returncode != 0:
                log(f"ERROR: Workload on {node_name} failed with code {proc.returncode}", RED, True)
                if stderr:
                    log(f"Error output from {node_name}:", RED, True)
                    print(stderr)
                if node_name != "local":
                    log(f"Full error log saved to: {error_output}", RED, True)
        
        log("All client workloads completed", GREEN, config["verbose"])
        
    else:
        # Single client mode (original behavior)
        run_command(cmd, output, config["log_file"])


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


def get_metrics(config, profile, thread):
    """
    Get metrics from output file(s).
    For multiple clients, aggregate results from all nodes.
    - Duration: print all nodes' values, then take minimum
    - Latency: print all nodes' values, then take minimum
    - Error ratio: aggregate all success/error counts
    """
    base_path = (
        f"{config['tar_dir']}/{config['wl_type']}-{config['wl_mode']}-"
        f"{config['db_combinations']}-{profile}-{thread}"
    )
    
    if config.get("multiple_clients", False):
        # Collect metrics from all nodes
        node_durations = {}
        node_latencies = {}
        total_success = 0
        total_error = 0
        
        node_names = ["local"] + client_nodes
        for node_name in node_names:
            node_suffix = f"_{node_name}" if node_name != "local" else "_local"
            file_path = f"{base_path}{node_suffix}.txt"
            
            if not os.path.exists(file_path):
                log(f"Warning: Result file not found for {node_name}: {file_path}", RED, True)
                continue
                
            with open(file_path, "r") as f:
                content = f.read()
            
            # Extract duration
            duration_match = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
            if duration_match:
                node_durations[node_name] = float(duration_match.group(1))
            
            # Extract latency
            latency_match = re.search(r"99th\(us\): ([0-9]+)", content)
            if latency_match:
                node_latencies[node_name] = int(latency_match.group(1))
            
            # Extract counts
            success_match = re.search(r"COMMIT .*?Count: ([0-9]+)", content)
            if success_match:
                total_success += int(success_match.group(1))
            
            if profile != "native":
                error_match = re.search(r"COMMIT_ERROR .*?Count: ([0-9]+)", content)
                if error_match:
                    total_error += int(error_match.group(1))
        
        # Print individual node metrics
        print(f"\n[{profile} thread={thread}] Node metrics:")
        print("  Duration (seconds):")
        for node_name in node_names:
            if node_name in node_durations:
                print(f"    {node_name}: {node_durations[node_name]:.4f}")
        
        print("  Latency p99 (us):")
        for node_name in node_names:
            if node_name in node_latencies:
                print(f"    {node_name}: {node_latencies[node_name]}")
        
        # Take minimum values
        min_duration = min(node_durations.values()) if node_durations else 0.0
        min_latency = min(node_latencies.values()) if node_latencies else 0
        
        print(f"  Selected minimum duration: {min_duration:.4f}")
        print(f"  Selected minimum latency: {min_latency}\n")
        
        # Calculate ratio from aggregated counts
        total = total_success + total_error
        ratio = total_error / total if total > 0 else 0.0
        
        return min_duration, min_latency, ratio
    
    else:
        # Single client mode (original behavior)
        file_path = f"{base_path}.txt"
        
        if not os.path.exists(file_path):
            return 0.0, 0, 0.0
        
        with open(file_path, "r") as f:
            content = f.read()

        # Extract duration
        duration_match = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
        duration = float(duration_match.group(1)) if duration_match else 0.0

        # Extract latency
        latency_match = re.search(r"99th\(us\): ([0-9]+)", content)
        latency = int(latency_match.group(1)) if latency_match else 0

        # Extract counts
        success_match = re.search(r"COMMIT .*?Count: ([0-9]+)", content)
        success_cnt = int(success_match.group(1)) if success_match else 0

        error_cnt = 0
        if profile != "native":
            error_match = re.search(r"COMMIT_ERROR .*?Count: ([0-9]+)", content)
            error_cnt = int(error_match.group(1)) if error_match else 0

        total = success_cnt + error_cnt
        ratio = error_cnt / total if total > 0 else 0.0

        return duration, latency, ratio


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
):
    print(f"{thread}:")
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
    cmd = f"echo '{password}' | sudo -S bash /root/oreo-ben/start-timeoracle.sh"
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

    # Change to script directory and go up one level
    os.chdir(Path(__file__).parent.parent)

    # Parse threads
    threads = (
        [int(t.strip()) for t in args.threads.split(",")]
        if args.threads
        else default_threads
    )
    
    # Validate threads for multiple-clients mode
    if args.multiple_clients:
        invalid_threads = [t for t in threads if t % 3 != 0]
        if invalid_threads:
            handle_error(
                f"In multiple-clients mode, all thread counts must be divisible by 3. "
                f"Invalid values: {invalid_threads}"
            )

    # Setup config
    wl_type = "ycsb"
    load_flag_dir = "./data/ycsb"
    tar_dir = "./data/ycsb"

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

    # Check config files
    if not Path(config["config_file"]).exists():
        handle_error(f"Config file {config['config_file']} does not exist")
    if not Path(config["bc_file"]).exists():
        handle_error(f"Config file {config['bc_file']} does not exist")

    # Build executable
    print("Building the benchmark executable")
    subprocess.run(["go", "build", "."])
    Path("cmd").rename("./bin/cmd")

    # Create directories
    Path(config["tar_dir"]).mkdir(parents=True, exist_ok=True)
    config["log_file"] = f"{config['tar_dir']}/benchmark.log"

    # Initialize results file
    mode_suffix = "_multi" if config["multiple_clients"] else ""
    results_file = (
        f"{config['tar_dir']}/{args.workload}_{args.db}_benchmark_results{mode_suffix}.csv"
    )
    with open(results_file, "w") as f:
        f.write(
            "thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err\n"
        )

    operation = extract_operation_count(config["config_file"])

    log(
        f"Running benchmark for [{wl_type}] workload with [{args.db}] database combinations",
        YELLOW,
        config["verbose"],
    )
    
    if config["multiple_clients"]:
        log("Multiple client mode enabled - will run on local, node5, and node6", YELLOW, True)
        log(f"Thread counts will be divided by 3 across nodes: {threads}", YELLOW, True)

    # Deployment
    if config["skip"]:
        log("Skipping deployment", YELLOW, config["verbose"])
    else:
        if config["remote"]:
            print("Running remotely")
            deploy_remote(config)
        else:
            print("Running locally")
            deploy_local(config)

    # Sync files to remote clients if multiple-clients mode is enabled
    if config["multiple_clients"]:
        log("Syncing files to client nodes...", YELLOW, True)
        
        # Test SSH connection first
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
        
        # Sync files
        for node in client_nodes:
            sync_files_to_remote(config, node)

    # Data loading check
    load_flag_file = Path(load_flag_dir) / f"{wl_type}-load"
    if config["loaded"]:
        log("Skipping data loading", YELLOW, config["verbose"])
    else:
        if load_flag_file.exists():
            log("Data has been already loaded", YELLOW, config["verbose"])
        else:
            log("Ready to load data", YELLOW, config["verbose"])

    # Prompt to continue
    # choice = input("Do you want to continue? (y/n): ").strip().lower()
    # if choice not in ["y", "yes"]:
    #     print("Exiting the script.")
    #     sys.exit(0)

    # Load data if needed
    if not load_flag_file.exists():
        load_data(config)

    # Run benchmarks
    for thread in threads:
        for profile in ["native", "cg", "oreo"]:
            output = (
                f"{config['tar_dir']}/{wl_type}-{args.workload}-"
                f"{args.db}-{profile}-{thread}.txt"
            )
            run_workload(config, "run", profile, thread, output)

        native_dur, native_p99, native_ratio = get_metrics(config, "native", thread)
        cg_dur, cg_p99, cg_ratio = get_metrics(config, "cg", thread)
        oreo_dur, oreo_p99, oreo_ratio = get_metrics(config, "oreo", thread)

        # Append results
        with open(results_file, "a") as f:
            f.write(
                f"{thread},{operation},{native_dur:.4f},{cg_dur:.4f},{oreo_dur:.4f},"
                f"{native_p99},{cg_p99},{oreo_p99},"
                f"{native_ratio:.4f},{cg_ratio:.4f},{oreo_ratio:.4f}\n"
            )

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
        )

        time.sleep(round_interval)

    cleanup(config)


if __name__ == "__main__":
    main()