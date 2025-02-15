
handle_error() {
    echo "Error: $1"
    exit 1
}

port=8001
db_combinations=
wl_mode=
verbose=false
limited=false
remove_all=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
    -p | --port)
        port="$2"
        shift
        ;;
    -wl | --workload)
        wl_mode="$2"    
        shift
        ;;
    -db | --db)
        db_combinations="$2"
        shift
        ;;
    -l | --limited) limited=true ;;
    -r | --remove) remove_all=true ;;
    -v | --verbose) verbose=true ;;
    *)
        echo "Unknown parameter passed: $1"
        exit 1
        ;;
    esac
    shift
done

if [ $wl_mode == "ycsb" ]; then
    bc=BenConfig_ycsb.yaml
else
    bc=BenConfig_realistic.yaml
fi

cd ~/oreo-ben || handle_error "oreo-ben directory not found"

ls ./config

if [ "$remove_all" = true ]; then
    docker rm -f executor-*
    exit 0
fi


extra_opts=""
if [ "$limited" = true ]; then
    extra_opts="--cpus=1"
fi

docker run --name="executor-$port" --network=host -d \
    -v ./config:/app/config \
    $extra_opts \
    oreo-executor \
    -p "$port" \
    -w "$wl_mode" \
    -bc "/app/config/$bc" \
    -db "$db_combinations"

sleep 3
    echo "Executor started"
    lsof -i ":$port"