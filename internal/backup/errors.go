package backup

import "fmt"

type BackupError struct {
	Op   string
	Path string
	Err  error
}

func (e *BackupError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
}

func newBackupError(op, path string, err error) error {
	return &BackupError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
