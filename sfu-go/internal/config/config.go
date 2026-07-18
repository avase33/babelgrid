package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr    string
	DSPURL  string
	GridURL string
	Workers int
	Targets []string
	Synth   bool
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	workers, _ := strconv.Atoi(getenv("BABELGRID_WORKERS", "4"))
	if workers < 1 {
		workers = 1
	}
	return Config{
		Addr:    getenv("BABELGRID_ADDR", ":8080"),
		DSPURL:  getenv("BABELGRID_DSP_URL", "http://localhost:8092"),
		GridURL: getenv("BABELGRID_GRID_URL", "http://localhost:8000"),
		Workers: workers,
		Targets: strings.Split(getenv("BABELGRID_TARGETS", "es,fr,de"), ","),
		Synth:   os.Getenv("BABELGRID_SYNTH") == "1",
	}
}
