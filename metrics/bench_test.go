package metrics_test

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/fairyhunter13/goresilience"
	"github.com/fairyhunter13/goresilience/bulkhead"
	"github.com/fairyhunter13/goresilience/circuitbreaker"
	"github.com/fairyhunter13/goresilience/metrics"
	"github.com/fairyhunter13/goresilience/retry"
	"github.com/fairyhunter13/goresilience/timeout"
)

var allokf = func(_ context.Context) error { return nil }

func BenchmarkMeasuredRunner(b *testing.B) {
	b.StopTimer()

	benchs := []struct {
		name    string
		wrapper func(r goresilience.Runner) goresilience.Runner
	}{
		{
			name: "Without measurement (Dummy).",
			wrapper: func(r goresilience.Runner) goresilience.Runner {
				return r
			},
		},
		{
			name: "With prometheus measurement.",
			wrapper: func(r goresilience.Runner) goresilience.Runner {
				promreg := prometheus.NewRegistry()
				rec := metrics.NewPrometheusRecorder(promreg)
				return metrics.NewMiddleware("bench", rec)(r)
			},
		},
	}

	for _, bench := range benchs {
		b.Run(bench.name, func(b *testing.B) {
			// Prepare the runner.
			runner := goresilience.RunnerChain(
				circuitbreaker.NewMiddleware(circuitbreaker.Config{}),
				bulkhead.NewMiddleware(bulkhead.Config{}),
				retry.NewMiddleware(retry.Config{}),
				timeout.NewMiddleware(timeout.Config{}))

			runner = bench.wrapper(runner)

			// execute the benhmark.
			for n := 0; n < b.N; n++ {
				b.StartTimer()
				runner.Run(context.TODO(), allokf)
				b.StopTimer()
			}
		})
	}
}
