package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileWriteJob represents a file to be written
type FileWriteJob struct {
	FilePath string
	Content  []byte
	Verbose  bool
}

// HighSpeedFileWriter handles bulk file writing operations
type HighSpeedFileWriter struct {
	jobs       chan FileWriteJob
	wg         sync.WaitGroup
	numWriters int
	dirCache   map[string]bool
	dirMutex   sync.RWMutex
}

// NewHighSpeedFileWriter creates a new high-speed file writer
func NewHighSpeedFileWriter(numWriters int) *HighSpeedFileWriter {
	fw := &HighSpeedFileWriter{
		jobs:       make(chan FileWriteJob, 10000), // Large buffer for batching
		numWriters: numWriters,
		dirCache:   make(map[string]bool),
	}

	// Start background writer workers
	for i := 0; i < numWriters; i++ {
		fw.wg.Add(1)
		go fw.worker(i)
	}

	return fw
}

// ensureDirFast creates directory with caching
func (fw *HighSpeedFileWriter) ensureDirFast(dir string) error {
	fw.dirMutex.RLock()
	if fw.dirCache[dir] {
		fw.dirMutex.RUnlock()
		return nil
	}
	fw.dirMutex.RUnlock()

	fw.dirMutex.Lock()
	defer fw.dirMutex.Unlock()

	if fw.dirCache[dir] {
		return nil
	}

	err := os.MkdirAll(dir, 0755)
	if err == nil {
		fw.dirCache[dir] = true
	}
	return err
}

// worker processes file write jobs
func (fw *HighSpeedFileWriter) worker(id int) {
	defer fw.wg.Done()

	for job := range fw.jobs {
		// Ensure directory exists
		dir := filepath.Dir(job.FilePath)
		if err := fw.ensureDirFast(dir); err != nil {
			if job.Verbose {
				fmt.Printf("Writer %d: Error creating dir %s: %v\n", id, dir, err)
			}
			continue
		}

		// Write file with large buffer
		file, err := os.Create(job.FilePath)
		if err != nil {
			if job.Verbose {
				fmt.Printf("Writer %d: Error creating file %s: %v\n", id, job.FilePath, err)
			}
			continue
		}

		// Use massive buffer for speed
		bufWriter := bufio.NewWriterSize(file, 1048576) // 1MB buffer
		_, err = bufWriter.Write(job.Content)
		if err != nil {
			file.Close()
			if job.Verbose {
				fmt.Printf("Writer %d: Error writing file %s: %v\n", id, job.FilePath, err)
			}
			continue
		}

		bufWriter.Flush()
		file.Close()

		if job.Verbose {
			fmt.Printf("🚀 Writer %d: Saved %s (%d bytes)\n", id, job.FilePath, len(job.Content))
		}
	}
}

// WriteFile queues a file for writing
func (fw *HighSpeedFileWriter) WriteFile(filePath string, content []byte, verbose bool) {
	select {
	case fw.jobs <- FileWriteJob{
		FilePath: filePath,
		Content:  content,
		Verbose:  verbose,
	}:
		// Job queued successfully
	default:
		// Channel full, write synchronously as fallback
		if verbose {
			fmt.Printf("Writer queue full, writing %s synchronously\n", filePath)
		}
		fw.writeSynchronous(filePath, content, verbose)
	}
}

// writeSynchronous writes file immediately
func (fw *HighSpeedFileWriter) writeSynchronous(filePath string, content []byte, verbose bool) {
	dir := filepath.Dir(filePath)
	fw.ensureDirFast(dir)

	file, err := os.Create(filePath)
	if err != nil {
		if verbose {
			fmt.Printf("Sync write error creating %s: %v\n", filePath, err)
		}
		return
	}
	defer file.Close()

	bufWriter := bufio.NewWriterSize(file, 1048576)
	bufWriter.Write(content)
	bufWriter.Flush()

	if verbose {
		fmt.Printf("Sync saved %s (%d bytes)\n", filePath, len(content))
	}
}

// Close shuts down the file writer
func (fw *HighSpeedFileWriter) Close() {
	close(fw.jobs)
	fw.wg.Wait()
}

// Global file writer instance
var globalFileWriter *HighSpeedFileWriter
var fileWriterOnce sync.Once

// GetFileWriter returns the global file writer instance
func GetFileWriter() *HighSpeedFileWriter {
	fileWriterOnce.Do(func() {
		globalFileWriter = NewHighSpeedFileWriter(16) // 16 dedicated file writers for maximum speed
		fmt.Println("🚀 High-speed file writer initialized with 16 workers")
	})
	return globalFileWriter
}
