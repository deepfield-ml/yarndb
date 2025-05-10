package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

// ConcurrentRead loads all YAML files concurrently
func (ds *YAMLDatastore) ConcurrentRead() error {
	// Collect YAML files
	var files []string
	err := filepath.Walk(ds.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			files = append(files, path)
			ds.files[path] = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Preallocate data map
	ds.mu.Lock()
	ds.data = make(map[string]interface{}, len(files)*100)
	ds.mu.Unlock()

	// Load files concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, len(files))
	start := time.Now()
	for _, path := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if err := ds.readFile(path); err != nil {
				errCh <- err
			}
			atomic.AddUint64(&ds.loadCount, 1)
		}(path)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(errCh)

	// Collect errors
	for err := range errCh {
		if err != nil {
			log.Errorf("Error reading file: %v", err)
		}
	}

	log.Infof("Loaded %d files in %v", atomic.LoadUint64(&ds.loadCount), time.Since(start))
	return nil
}

// readFile reads and parses a single YAML file
func (ds *YAMLDatastore) readFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	var fileData map[string]interface{}
	if err := yaml.Unmarshal(data, &fileData); err != nil {
		return err
	}

	// Merge into global cache
	ds.mu.Lock()
	defer ds.mu.Unlock()
	for id, record := range fileData {
		ds.data[id] = record
		ds.updateIndexes(id, record)
	}
	log.Debugf("Loaded file %s with %d records", path, len(fileData))
	return nil
}
