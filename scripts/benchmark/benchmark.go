package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pterm/pterm"
)

var (
	successCount int64
	errorCount   int64
	requestID    uint64
)

func main() {
	// CLI Parameters
	targetURL := flag.String("url", "http://localhost:3000/v1/collection", "Target URL")
	concurrency := flag.Int("c", 100, "Workers")
	duration := flag.Duration("d", 30*time.Second, "Duration")
	method := flag.String("m", "POST", "Method")
	auth := flag.String("auth", "Bearer benchmark-secret-key", "Auth Header")
	flag.Parse()

	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithMargin(10).Println("MOCKSERVER PERFORMANCE BENCHMARK")
	fmt.Printf("Target  : %s [%s]\nWorkers : %d\nDuration: %v\n\n", *targetURL, *method, *concurrency, *duration)

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	// HTTP Client Optimization
	transport := &http.Transport{
		MaxIdleConns:        *concurrency,
		MaxIdleConnsPerHost: *concurrency,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableCompression:  true,
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	results := make(chan time.Duration, 1000000)
	var wg sync.WaitGroup
	wg.Add(*concurrency)

	start := time.Now()

	liveArea, _ := pterm.DefaultArea.Start()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				elapsed := time.Since(start).Seconds()
				currSuccess := atomic.LoadInt64(&successCount)
				currErrors := atomic.LoadInt64(&errorCount)
				rps := float64(currSuccess+currErrors) / elapsed

				stats, _ := pterm.DefaultTable.WithData(pterm.TableData{
					{"Current RPS", "Success", "Errors", "Elapsed"},
					{fmt.Sprintf("%.2f", rps), pterm.FgGreen.Sprintf("%d", currSuccess), pterm.FgRed.Sprintf("%d", currErrors), fmt.Sprintf("%.1fs", elapsed)},
				}).Srender()

				liveArea.Update(stats)
			}
		}
	}()

	// WORKERS
	for i := 0; i < *concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					reqStart := time.Now()

					var bodyBuffer *bytes.Buffer
					if *method == "POST" || *method == "PUT" {
						id := atomic.AddUint64(&requestID, 1)
						payload := `{"id": "bench_` + strconv.FormatUint(id, 10) + `", "data": "benchmark"}`
						bodyBuffer = bytes.NewBufferString(payload)
					} else {
						bodyBuffer = bytes.NewBuffer(nil)
					}

					req, _ := http.NewRequestWithContext(ctx, *method, *targetURL, bodyBuffer)
					req.Header.Set("Authorization", *auth)
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Connection", "keep-alive")

					resp, err := client.Do(req)
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						continue
					}

					if resp.StatusCode >= 400 {
						atomic.AddInt64(&errorCount, 1)
					} else {
						atomic.AddInt64(&successCount, 1)
						select {
						case results <- time.Since(reqStart):
						default:
						}
					}
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()
	liveArea.Stop()
	close(results)
	totalDuration := time.Since(start)

	generateFinalReport(results, successCount, errorCount, totalDuration)
}

func generateFinalReport(results chan time.Duration, success, errors int64, totalDur time.Duration) {
	var latencies []time.Duration
	for l := range results {
		latencies = append(latencies, l)
	}

	totalReq := success + errors
	if totalReq == 0 {
		pterm.Error.Println("No requests completed.")
		return
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p50 := latencies[int(float64(len(latencies))*0.50)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	rps := float64(totalReq) / totalDur.Seconds()

	pterm.DefaultSection.Println("FINAL BENCHMARK RESULTS")

	tableData := pterm.TableData{
		{"Metric", "Value"},
		{"Throughput (RPS)", pterm.FgCyan.Sprintf("%.2f req/s", rps)},
		{"Total Requests", fmt.Sprintf("%d", totalReq)},
		{"Success Rate", pterm.FgGreen.Sprintf("%.2f%%", float64(success)/float64(totalReq)*100)},
		{"P50 (Median)", fmt.Sprintf("%v", p50.Round(time.Microsecond))},
		{"P95 (Tail)", pterm.FgYellow.Sprintf("%v", p95.Round(time.Microsecond))},
		{"Total Errors", pterm.FgRed.Sprintf("%d", errors)},
	}

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
}

func avg(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	var sum int64
	for _, l := range latencies {
		sum += l.Nanoseconds()
	}
	return time.Duration(sum / int64(len(latencies)))
}
