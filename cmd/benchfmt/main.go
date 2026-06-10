package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
)

func main() {
	count := flag.Int("count", 1, "run each benchmark n times")
	timeout := flag.String("timeout", "300s", "timeout for the benchmark run")
	noMem := flag.Bool("nomem", false, "disable memory allocation reporting")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go run ./cmd/benchfmt/ [flags] [filter]\n\n")
		fmt.Fprintf(os.Stderr, "Runs benchmarks in cmd/bench/ and formats the output.\n")
		fmt.Fprintf(os.Stderr, "The optional filter argument is a regex passed to -bench (default \".\").\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	filter := "."
	if flag.NArg() > 0 {
		filter = flag.Arg(0)
	}

	args := []string{"test",
		"-bench=" + filter,
		"-count=" + strconv.Itoa(*count),
		"-timeout", *timeout,
		"./cmd/bench/",
	}
	if !*noMem {
		args = append(args[:3], append([]string{"-benchmem"}, args[3:]...)...)
	}

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(tw, "BENCHMARK\tITERS\tTIME/OP\tTHROUGHPUT\tMEMORY/OP\tALLOCS/OP\n")
	fmt.Fprintf(tw, "---------\t-----\t-------\t----------\t---------\t---------\n")

	var total int
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		iters := fields[1]

		var timeOp, throughput, memOp, allocsOp string

		for i := 2; i < len(fields)-1; i++ {
			switch fields[i+1] {
			case "ns/op":
				timeOp = formatTime(fields[i])
			case "MB/s":
				throughput = fields[i] + " MB/s"
			case "B/op":
				memOp = formatBytes(fields[i])
			case "allocs/op":
				allocsOp = fields[i]
			}
		}

		if throughput == "" || throughput == "0.00 MB/s" {
			throughput = "-"
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			name, iters, timeOp, throughput, memOp, allocsOp)
		total++
	}

	tw.Flush()

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "\nbenchmark run failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%d benchmarks completed.\n", total)
}

func formatTime(nsStr string) string {
	var ns float64
	fmt.Sscanf(nsStr, "%f", &ns)

	switch {
	case ns >= 1_000_000_000:
		return fmt.Sprintf("%.2fs", ns/1_000_000_000)
	case ns >= 1_000_000:
		return fmt.Sprintf("%.1fms", ns/1_000_000)
	case ns >= 1_000:
		return fmt.Sprintf("%.1fus", ns/1_000)
	default:
		return fmt.Sprintf("%.1fns", ns)
	}
}

func formatBytes(bStr string) string {
	var b float64
	fmt.Sscanf(bStr, "%f", &b)

	switch {
	case b >= 1_073_741_824:
		return fmt.Sprintf("%.1f GB", b/1_073_741_824)
	case b >= 1_048_576:
		return fmt.Sprintf("%.1f MB", b/1_048_576)
	case b >= 1_024:
		return fmt.Sprintf("%.1f KB", b/1_024)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}
