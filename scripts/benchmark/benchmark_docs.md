# Start Benchmark Mockserver YAML config file (slient)

### With Slient
This is recommended for better and more accurate results.

```bash
 go run . start --config ./scripts/benchmark/mockserver.benchmark.yaml > NUL
```

### Normal

```bash
 go run . start --config ./scripts/mockserver.benchmark.yaml
```



# Start Test Any Endpoint
Before starting the performance test, make sure you have started Mockserver.

```bash
go run scripts//benchmark/benchmark.go -url http://localhost:3000/v1/health -m GET -c 100 -d 15s
```

- **-url:** URL where the test will be performed 
- **-c:** total worker
- **-d:** how long should it last duration