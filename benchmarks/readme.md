## Start a benchmark

```bash
# load data
# 2 indicates 2 threads
go run main.go redis load 2
# test redids
# 2 indicates 2 threads
go run main.go redis run 2
```