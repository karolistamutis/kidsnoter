package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/karolistamutis/kidsnoter/cmd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Set up Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":9091", nil); err != nil {
			fmt.Printf("Error starting metrics server: %v\n", err)
		}
	}()

	// Execute the root command
	if err := cmd.RootCmd.Execute(); err != nil {
		if !errors.Is(err, cmd.ErrSilent) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}
