package main

// Config holds application configuration
type Config struct {
    Debug    bool
    Verbose  bool
    Output   string
    LogLevel int
    Timeout  int
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
    return Config{
        Debug:    false,
        Verbose:  false,
        Output:   "stdout",
        LogLevel: 1,
        Timeout:  30,
    }
}

// NewConfig creates a new configuration with custom values
func NewConfig(debug bool, output string) Config {
    cfg := DefaultConfig()
    cfg.Debug = debug
    cfg.Output = output
    return cfg
}
