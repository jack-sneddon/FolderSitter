package backup

import (
	"os"
	"path/filepath"
)

// Contains createTasks() and related task management methods
// createTasks generates the list of files to be backed up
func (s *Service) createTasks() ([]CopyTask, error) {
	var tasks []CopyTask

	for _, folder := range s.config.FoldersToBackup {
		srcPath := filepath.Join(s.config.SourceDirectory, folder)
		dstPath := filepath.Join(s.config.TargetDirectory, folder)

		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip if matches exclude patterns
			for _, pattern := range s.config.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					s.logger.Debug("Skipping excluded file: %s", path)
					return nil
				}
			}

			// Create relative path
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}

			destPath := filepath.Join(dstPath, relPath)

			if !info.IsDir() {
				tasks = append(tasks, CopyTask{
					Source:      path,
					Destination: destPath,
					Size:        info.Size(),
					ModTime:     info.ModTime(),
				})
			}

			return nil
		})

		if err != nil {
			return nil, newBackupError("CreateTasks", srcPath, err)
		}
	}

	return tasks, nil
}
