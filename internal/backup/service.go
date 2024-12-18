// service.go
package backup

import "fmt"

// NewService creates a new backup service instance
// service.go
func NewService(cfg *Config) (*Service, error) {
	logger, err := NewLogger(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	versioner, err := NewVersionManager(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create version manager: %v", err)
	}

	s := &Service{
		config:    cfg,
		logger:    logger,
		versioner: versioner,
	}

	s.pool = NewWorkerPool(
		cfg.Concurrency,
		s.copyFile,
		cfg.RetryAttempts,
		cfg.RetryDelay,
	)

	return s, nil
}

// Version management methods
func (s *Service) GetVersions() []BackupVersion {
	if s.versioner == nil {
		return nil
	}
	return s.versioner.GetVersions()
}

func (s *Service) GetVersion(id string) (*BackupVersion, error) {
	if s.versioner == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	return s.versioner.GetVersion(id)
}

func (s *Service) GetLatestVersion() (*BackupVersion, error) {
	if s.versioner == nil {
		return nil, fmt.Errorf("version manager not initialized")
	}
	latest := s.versioner.GetLatestVersion()
	if latest == nil {
		return nil, fmt.Errorf("no backup versions found")
	}
	return latest, nil
}
