// service.go
package backup

import "fmt"

// NewService creates a new backup service instance
func NewService(cfg *Config) (*Service, error) {
	logger, err := NewLogger(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// Validate configuration before creating service
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	s := &Service{
		config:  cfg,
		logger:  logger,
		metrics: &Metrics{},
	}

	s.pool = NewWorkerPool(
		cfg.Concurrency,
		s.copyFile,
		cfg.RetryAttempts,
		cfg.RetryDelay,
	)
	return s, nil
}
