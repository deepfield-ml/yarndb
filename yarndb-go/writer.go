package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Save writes the in-memory state to disk
func (ds *YAMLDatastore) Save() error {
	ds.saveMu.Lock()
	defer ds.saveMu.Unlock()

	ds.mu.RLock()
	if !ds.dirty {
		ds.mu.RUnlock()
		return nil
	}
	dataCopy := make(map[string]interface{}, len(ds.data))
	for id, data := range ds.data {
		dataCopy[id] = data
	}
	filesCopy := make(map[string]bool, len(ds.files))
	for file := range ds.files {
		filesCopy[file] = true
	}
	ds.mu.RUnlock()

	// Group records by file
	fileData := make(map[string]map[string]interface{})
	for path := range filesCopy {
		fileData[path] = make(map[string]interface{})
	}
	for id, data := range dataCopy {
		fileID := strings.Split(id, "_")[0]
		filePath := filepath.Join(ds.dir, "records_"+fileID+".yaml")
		if _, exists := filesCopy[filePath]; !exists {
			filePath = filepath.Join(ds.dir, "records_default.yaml")
		}
		fileData[filePath][id] = data
	}

	// Write each file
	for path, records := range fileData {
		if len(records) == 0 {
			continue
		}
		out, err := yaml.Marshal(records)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, out, 0644); err != nil {
			return err
		}
	}

	ds.mu.Lock()
	ds.dirty = false
	ds.mu.Unlock()

	log.Info("Saved state to disk")
	return nil
}
