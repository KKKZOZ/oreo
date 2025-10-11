#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse
import os
import re
import subprocess
import sys
import time
from pathlib import Path

# ===== Terminal color codes =====
RED = "\033[0;31m"
GREEN = "\033[0;32m"
YELLOW = "\033[1;33m"
BLUE = "\033[0;34m"
PURPLE = "\033[0;35m"
CYAN = "\033[0;36m"
NC = "\033[0m"

# ===== Global config (mirrors the bash script) =====
executor_port = 8001
timeoracle_port = 8010
thread_load = 50
default_threads = [12, 24, 36, 48, 60, 84, 96]
round_interval = 5

# Default nodes (will be trimmed/expanded for remote depending on wl)
node_list_default = ["node2", "node3", "node4", "node5"]
client_nodes = ["node6", "node5"]  # for --multiple-clients mode (local + 2 clients)

password = "kkkzoz"  # used for remote sudo scripts

executor_pid = None
timeoracle_pid = None


# ===== Utils =====
def log(msg, color=NC, verbose=True):
    if verbose:
        print(f"{color}{msg}{NC}")


def handle_error(msg):
    print(f"Error: {msg}")
    sys.exit(1)


def kill_process_on_port(port):
    try:
        result = subprocess.run(["lsof", "-t", "-i", f":{port}"], capture_output=True, text=True)
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


# ===== Remote sync (for multi-clients mode) =====
def sync_files_to_remote(config, node):
    """
    Sync binary and config files to remote node for multi-client execution.
    Layout on remote: /tmp/oreo-benchmark/{bin,workloads/realistic,config}
    """
    log(f"Syncing files to {node}", CYAN, True)
    remote_base = "/tmp/oreo-benchmark"

    try:
        subprocess.run(
            ["ssh", node, f"mkdir -p {remote_base}/bin {remote_base}/workloads/realistic {remote_base}/config"],
            capture_output=True,
            text=True,
            check=True,
        )

        # Sync binary
        log(f"  Syncing binary to {node}...", CYAN, True)
        subprocess.run(["rsync", "-az", "./bin/cmd", f"{node}:{remote_base}/bin/"], check=True, capture_output=True)

        # Sync config files
        log(f"  Syncing config files to {node}...", CYAN, True)
        subprocess.run(
            ["rsync", "-az", config["config_file"], f"{node}:{remote_base}/workloads/realistic/"],
            check=True,
            capture_output=True,
        )
        subprocess.run(
            ["rsync", "-az", config["bc_file"], f"{node}:{remote_base}/config/"],
            check=True,
            capture_output=True,
        )

        # Verify
        verify_cmd = f"ls -lh {remote_base}/bin/cmd {remote_base}/workloads/realistic/ {remote_base}/config/"
        result = subprocess.run(["ssh", node, verify_cmd], capture_output=True, text=True, check=True)
        log(f"  Files on {node}:\n{result.stdout}", GREEN, True)

        log(f"Files synced to {node} successfully", GREEN, True)
    except subprocess.CalledProcessError as e:
        log(f"ERROR: Failed to sync files to {node}", RED, True)
        log(f"Command: {e.cmd}", RED, True)
        log(f"Return code: {e.returncode}", RED, True)
        if e.stdout:
            log(f"stdout: {e.stdout}", RED, True)
        if e.stderr:
            log(f"stderr: {e.stderr}", RED, True)
        raise


def run_workload_remote(node, config, mode, profile, thread):
    """
    Run workload on a remote node (used in --multiple-clients mode).
    """
    remote_base = "/tmp/oreo-benchmark"
    config_filename = os.path.basename(config["config_file"])
    bc_filename = os.path.basename(config["bc_file"])

    cmd = (
        f"cd {remote_base} && ./bin/cmd "
        f"-d oreo "
        f"-wl {config['wl_type']} "
        f"-wc ./workloads/realistic/{config_filename} "
        f"-bc ./config/{bc_filename} "
        f"-m {mode} "
        f"-ps {profile} "
        f"-t {thread}"
    )
    log(f"Running workload on {node}: {profile} thread={thread}", BLUE, True)
    proc = subprocess.Popen(["ssh", node, cmd], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    return proc


# ===== Core workload runner =====
def run_workload(config, mode, profile, thread, output):
    """
    Run a single (profile, thread) workload, in either single-client or multi-clients mode.
    """
    cmd_base = [
        "./bin/cmd",
        "-d", "oreo",
        "-wl", config["wl_type"],
        "-wc", config["config_file"],
        "-bc", config["bc_file"],
        "-m", mode,
        "-ps", profile,
    ]

    if config.get("multiple_clients", False):
        total_nodes = 1 + len(client_nodes)
        thread_per_node = thread // total_nodes
        log(
            f"Running {config['wl_type']} {profile} total_thread={thread} "
            f"(per-node={thread_per_node}, nodes={total_nodes})",
            BLUE,
            True,
        )

        processes = []

        # local
        local_output = output.replace(".txt", "_local.txt")
        local_cmd = cmd_base + ["-t", str(thread_per_node)]
        log(f"Starting LOCAL workload with {thread_per_node} threads", BLUE, True)
        local_proc = subprocess.Popen(
            local_cmd,
            stdout=open(local_output, "w"),
            stderr=open(config["log_file"], "a") if config["log_file"] else subprocess.DEVNULL,
        )
        processes.append(("local", local_proc, local_output))

        # remote clients
        for node in client_nodes:
            node_output = output.replace(".txt", f"_{node}.txt")
            remote_proc = run_workload_remote(node, config, mode, profile, thread_per_node)
            processes.append((node, remote_proc, node_output))

        # wait / collect
        log("Waiting for all client workloads to complete...", YELLOW, True)
        for node_name, proc, node_output in processes:
            if node_name == "local":
                ret = proc.wait()
                if ret != 0:
                    log(f"ERROR: Local workload failed with code {ret}", RED, True)
                continue

            stdout, stderr = proc.communicate()
            with open(node_output, "w") as f:
                f.write(stdout or "")
            if stderr:
                err_file = node_output.replace(".txt", "_error.log")
                with open(err_file, "w") as f:
                    f.write(stderr)
                if proc.returncode != 0:
                    log(f"ERROR: Remote workload on {node_name} failed ({proc.returncode}), see {err_file}", RED, True)

        log("All client workloads completed", GREEN, True)

    else:
        # single client
        cmd = cmd_base + ["-t", str(thread)]
        run_command(cmd, output, config["log_file"])


# ===== Data loading =====
def load_data(config):
    for profile in ["native", "cg", "oreo"]:
        log(f"Loading to {config['wl_type']} {profile}", BLUE, True)
        cmd = [
            "./bin/cmd",
            "-d", "oreo",
            "-wl", config["wl_type"],
            "-wc", config["config_file"],
            "-bc", config["bc_file"],
            "-m", "load",
            "-ps", profile,
            "-t", str(config["thread_load"]),
        ]
        subprocess.run(cmd)

    load_flag = Path(config["tar_dir"]) / f"{config['wl_type']}-load"
    load_flag.touch()


# ===== Metrics parsing =====
def get_metrics(config, profile, thread):
    base = f"{config['tar_dir']}/{config['wl_type']}-{profile}-{thread}"

    if config.get("multiple_clients", False):
        node_names = ["local"] + client_nodes
        durations = {}
        p99s = {}
        succ_total = 0
        err_total = 0

        for name in node_names:
            suffix = "_local" if name == "local" else f"_{name}"
            fpath = f"{base}{suffix}.txt"
            if not os.path.exists(fpath):
                log(f"Warning: missing result file: {fpath}", YELLOW, True)
                continue
            with open(fpath, "r") as f:
                content = f.read()

            m = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
            if m:
                durations[name] = float(m.group(1))

            m = re.search(r"99th\(us\):\s*([0-9]+)", content)
            if m:
                p99s[name] = int(m.group(1))

            m = re.search(r"COMMIT\s+.*?Count:\s*([0-9]+)", content)
            if m:
                succ_total += int(m.group(1))
            if profile != "native":
                m = re.search(r"COMMIT_ERROR\s+.*?Count:\s*([0-9]+)", content)
                if m:
                    err_total += int(m.group(1))

        avg_dur = (sum(durations.values()) / len(durations)) if durations else 0.0
        max_p99 = max(p99s.values()) if p99s else 0
        total = succ_total + err_total
        ratio = (err_total / total) if total > 0 else 0.0

        print(f"\n[{profile} thread={thread}] Node metrics:")
        if durations:
            print("  Duration (s):")
            for n, v in durations.items():
                print(f"    {n}: {v:.4f}")
        if p99s:
            print("  Latency p99 (us):")
            for n, v in p99s.items():
                print(f"    {n}: {v}")
        print(f"  Selected average duration: {avg_dur:.4f}")
        print(f"  Selected maximum latency: {max_p99}\n")

        return avg_dur, max_p99, ratio

    else:
        fpath = f"{base}.txt"
        if not os.path.exists(fpath):
            return 0.0, 0, 0.0
        with open(fpath, "r") as f:
            content = f.read()

        m = re.search(r"^Run finished.*?([0-9.]+)", content, re.MULTILINE)
        dur = float(m.group(1)) if m else 0.0
        m = re.search(r"99th\(us\):\s*([0-9]+)", content)
        p99 = int(m.group(1)) if m else 0

        m = re.search(r"COMMIT\s+.*?Count:\s*([0-9]+)", content)
        succ = int(m.group(1)) if m else 0
        err = 0
        if profile != "native":
            m = re.search(r"COMMIT_ERROR\s+.*?Count:\s*([0-9]+)", content)
            err = int(m.group(1)) if m else 0
        total = succ + err
        ratio = (err / total) if total > 0 else 0.0
        return dur, p99, ratio


def print_summary(thread, nd, cd, od, np, cp, op, nr, cr, or_):
    print(f"{thread}:")
    print(f"native:{nd:.4f}")
    print(f"cg    :{cd:.4f}")
    print(f"oreo  :{od:.4f}")
    rn = (od / nd) if nd > 0 else 0.0
    rcg = (od / cd) if cd > 0 else 0.0
    print(f"Oreo:native = {rn:.5f}")
    print(f"Oreo:cg     = {rcg:.5f}")
    print(f"native 99th: {np}")
    print(f"cg     99th: {cp}")
    print(f"oreo   99th: {op}")
    print("Error ratio:")
    print(f"native = {nr:.4f}")
    print(f"cg = {cr:.4f}")
    print(f"oreo = {or_:.4f}")
    print("---------------------------------")


# ===== Local/Remote deploy & cleanup =====
def deploy_local(config):
    global executor_pid, timeoracle_pid
    kill_process_on_port(executor_port)
    kill_process_on_port(timeoracle_port)

    log("Starting executor", GREEN, True)
    # MUST pass -db always
    cmd = [
        "./bin/executor",
        "-p", str(executor_port),
        "-w", config["wl_type"],
        "-bc", config["bc_file"],
        "-db", config["db_combinations"],
    ]
    executor_log = open("./log/executor.log", "w")
    proc = subprocess.Popen(cmd, stderr=executor_log)
    executor_pid = proc.pid

    log("Starting time oracle", GREEN, True)
    timeoracle_log = open("./log/timeoracle.log", "w")
    proc2 = subprocess.Popen(
        ["./bin/timeoracle", "-p", str(timeoracle_port), "-type", "hybrid"],
        stdout=subprocess.DEVNULL,
        stderr=timeoracle_log,
    )
    timeoracle_pid = proc2.pid


def deploy_remote(config):
    # Adjust node list like bash
    node_list = list(node_list_default)
    if config["wl_type"] == "iot":
        node_list = ["node2", "node3"]
    elif config["wl_type"] == "social":
        node_list = ["node2", "node3", "node4"]

    log("Setup timeoracle on first node", GREEN, True)
    cmd = f"echo '{password}' | sudo -S bash /root/oreo-ben/start-timeoracle.sh"
    subprocess.run(["ssh", "-t", node_list[0], cmd])

    for node in node_list:
        log(f"Setup {node}", GREEN, True)
        # -wl ycsb: we explicitly pass db_combinations
        remote_cmd = (
            f"echo '{password}' | sudo -S bash /root/oreo-ben/start-ft-executor-docker.sh "
            f"-p {executor_port} -wl ycsb -bc BenConfig_realistic.yaml -db {config['db_combinations']} -r"
        )
        subprocess.run(["ssh", "-t", node, remote_cmd])


def cleanup(config):
    if not config["remote"]:
        if executor_pid:
            log("Killing executor", RED, True)
            subprocess.run(["kill", str(executor_pid)])
        if timeoracle_pid:
            log("Killing time oracle", RED, True)
            subprocess.run(["kill", str(timeoracle_pid)])


# ===== Helpers =====
def extract_operation_count(config_file):
    try:
        with open(config_file, "r") as f:
            for line in f:
                m = re.match(r"^operationcount.*?([0-9.]+)", line.strip())
                if m:
                    return m.group(1)
    except FileNotFoundError:
        return "0"
    return "0"


def parse_args():
    p = argparse.ArgumentParser(description="Realistic workload benchmark (pythonized from bash)")
    # Only allow these 3 workloads per your rule
    p.add_argument("-wl", "--workload", required=True, choices=["iot", "hotel", "social"], help="Workload type")
    p.add_argument("-t", "--threads", help="Thread counts (comma-separated), default: 8,16,32,48,64,80,96")
    p.add_argument("-v", "--verbose", action="store_true", help="Verbose log")
    p.add_argument("-r", "--remote", action="store_true", help="Deploy executors/timeoracle remotely")
    p.add_argument("-s", "--skip", action="store_true", help="Skip deployment (use existing services)")
    p.add_argument("-l", "--loaded", action="store_true", help="Assume data is already loaded")
    p.add_argument(
        "-mc", "--multiple-clients", action="store_true",
        help="Run workload on multiple client nodes (local + node5 + node6) and aggregate metrics"
    )
    # NOTE: --db is intentionally removed per requirement
    return p.parse_args()


def main():
    args = parse_args()

    # cd to repo root (same as bash: script root then ..)
    os.chdir(Path(__file__).parent)
    if (Path.cwd() / "workloads").exists() is False and (Path.cwd().parent / "workloads").exists():
        os.chdir(Path.cwd().parent)

    # parse threads
    threads = default_threads if not args.threads else [int(x.strip()) for x in args.threads.split(",") if x.strip()]

    # in multi-clients mode, require divisibility by (1 + len(client_nodes))
    if args.multiple_clients:
        total_nodes = 1 + len(client_nodes)
        invalid = [t for t in threads if t % total_nodes != 0]
        if invalid:
            handle_error(
                f"In --multiple-clients mode, threads must be divisible by {total_nodes}. Invalid: {invalid}"
            )

    # Determine db_combinations by wl
    wl_type = args.workload
    db_map = {
        "iot": "Redis,MongoDB2",
        "hotel": "Redis,MongoDB2,Cassandra",
        "social": "Redis,KVRocks,MongoDB2,Cassandra",
    }
    if wl_type not in db_map:
        handle_error(f"Unsupported workload '{wl_type}' for db_combinations mapping")
    db_combinations = db_map[wl_type]

    # config files/dirs
    bc_file = "./config/BenConfig_realistic.yaml"
    config_file = f"./workloads/realistic/{wl_type}.yaml"
    tar_dir = f"./data/{wl_type}"
    Path(tar_dir).mkdir(parents=True, exist_ok=True)
    Path("./log").mkdir(parents=True, exist_ok=True)

    # build
    print("Building the benchmark")
    subprocess.run(["go", "build", "."])
    Path("./bin").mkdir(exist_ok=True)
    if Path("cmd").exists():
        Path("cmd").rename("./bin/cmd")

    # config dict
    config = {
        "wl_type": wl_type,
        "bc_file": bc_file,
        "config_file": config_file,
        "tar_dir": tar_dir,
        "log_file": f"{tar_dir}/benchmark.log",
        "thread_load": thread_load,
        "remote": args.remote,
        "skip": args.skip,
        "loaded": args.loaded,
        "multiple_clients": args.multiple_clients,
        "db_combinations": db_combinations,  # <- from mapping
    }

    # sanity checks
    if not Path(config_file).exists():
        handle_error(f"Config file {config_file} does not exist")
    if not Path(bc_file).exists():
        handle_error(f"Config file {bc_file} does not exist")

    # results file
    results_file = f"{tar_dir}/{wl_type}_benchmark_results.csv"
    with open(results_file, "w") as f:
        f.write("thread,operation,native,cg,oreo,native_p99,cg_p99,oreo_p99,native_err,cg_err,oreo_err\n")

    # operation count
    operation = extract_operation_count(config_file)

    # deployment
    if config["skip"]:
        log("Skipping deployment", YELLOW, True)
    else:
        if config["remote"]:
            print("Running remotely")
            deploy_remote(config)
        else:
            print("Running locally")
            deploy_local(config)

    # data loading
    load_flag = Path(tar_dir) / f"{wl_type}-load"
    if config["loaded"]:
        log("Skipping data loading", YELLOW, True)
    else:
        if load_flag.exists():
            log("Data has been already loaded", YELLOW, True)
        else:
            log("Ready to load data", YELLOW, True)
            load_data(config)

    # multi-clients: check & sync
    if config["multiple_clients"]:
        log("Multiple client mode enabled - local + node5 + node6", YELLOW, True)
        for node in client_nodes:
            log(f"Testing SSH to {node}...", CYAN, True)
            try:
                subprocess.run(
                    ["ssh", "-o", "ConnectTimeout=5", node, "echo 'Connection OK'"],
                    capture_output=True, text=True, check=True, timeout=10
                )
                log(f"  Connection to {node}: OK", GREEN, True)
            except subprocess.TimeoutExpired:
                handle_error(f"SSH connection to {node} timed out")
            except subprocess.CalledProcessError as e:
                handle_error(f"Cannot connect to {node}: {e.stderr}")
        for node in client_nodes:
            sync_files_to_remote(config, node)

    # run workloads
    for thread in threads:
        for profile in ["native", "cg", "oreo"]:
            output = f"{tar_dir}/{wl_type}-{profile}-{thread}.txt"
            run_workload(config, "run", profile, thread, output)

        native_dur, native_p99, native_ratio = get_metrics(config, "native", thread)
        cg_dur, cg_p99, cg_ratio = get_metrics(config, "cg", thread)
        oreo_dur, oreo_p99, oreo_ratio = get_metrics(config, "oreo", thread)

        with open(results_file, "a") as f:
            f.write(
                f"{thread},{operation},{native_dur:.4f},{cg_dur:.4f},{oreo_dur:.4f},"
                f"{native_p99},{cg_p99},{oreo_p99},{native_ratio:.4f},{cg_ratio:.4f},{oreo_ratio:.4f}\n"
            )

        print_summary(
            thread,
            native_dur, cg_dur, oreo_dur,
            native_p99, cg_p99, oreo_p99,
            native_ratio, cg_ratio, oreo_ratio,
        )

        time.sleep(round_interval)

    cleanup(config)


if __name__ == "__main__":
    main()
