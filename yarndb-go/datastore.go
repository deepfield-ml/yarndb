package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"time"

	"github.com/spf13/viper"
)

// YAMLDatastore manages generic YAML data in memory
type YAMLDatastore struct {
	mu        sync.RWMutex
	data      map[string]interface{}            // In-memory cache: recordID -> data
	files     map[string]bool                   // Track YAML files
	indexes   map[string]map[interface{}]string // Indexes: key -> value -> recordID
	dir       string                            // Directory containing YAML files
	cache     map[string]interface{}            // Cache for merged data
	cacheLock sync.RWMutex                      // Lock for cache
	dirty     bool                              // Flag for unsaved changes
	saveMu    sync.Mutex                        // Lock for file writes
	loadCount uint64                            // Atomic counter for loaded files
	txMu      sync.Mutex                        // Lock for transactions
	txActive  bool                              // Flag for active transaction
	txData    map[string]interface{}            // Transaction data
	txFiles   map[string]bool                   // Transaction files
}

// NewYAMLDatastore initializes the datastore and starts auto-save
func NewYAMLDatastore(dir string) (*YAMLDatastore, error) {
	ds := &YAMLDatastore{
		data:      make(map[string]interface{}, 1000),
		files:     make(map[string]bool, 100),
		indexes:   make(map[string]map[interface{}]string),
		dir:       dir,
		cache:     make(map[string]interface{}),
		dirty:     false,
		loadCount: 0,
		txActive:  false,
		txData:    make(map[string]interface{}),
		txFiles:   make(map[string]bool),
	}

	// Load all YAML files concurrently
	if err := ds.ConcurrentRead(); err != nil {
		return nil, err
	}

	// Start auto-save goroutine
	go ds.autoSave()

	// Handle OS signals for graceful shutdown
	go ds.handleSignals()

	return ds, nil
}

// Set creates or updates a record
func (ds *YAMLDatastore) Set(recordID string, data interface{}, fileID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.txActive {
		ds.txData[recordID] = data
		ds.txFiles[filepath.Join(ds.dir, "records_"+fileID+".yaml")] = true
		return nil
	}

	ds.data[recordID] = data
	ds.files[filepath.Join(ds.dir, "records_"+fileID+".yaml")] = true
	ds.dirty = true
	ds.invalidateCache()
	ds.updateIndexes(recordID, data)
	log.Infof("Set record %s in file %s", recordID, fileID)
	return nil
}

// Get retrieves a record by ID
func (ds *YAMLDatastore) Get(recordID string) (interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if ds.txActive {
		if data, exists := ds.txData[recordID]; exists {
			return data, nil
		}
	}
	data, exists := ds.data[recordID]
	if !exists {
		return nil, nil
	}
	log.Debugf("Retrieved record %s", recordID)
	return data, nil
}

// Delete removes a record
func (ds *YAMLDatastore) Delete(recordID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.txActive {
		ds.txData[recordID] = nil
		return nil
	}

	if _, exists := ds.data[recordID]; !exists {
		return errors.New("record not found")
	}
	delete(ds.data, recordID)
	ds.dirty = true
	ds.invalidateCache()
	for key, index := range ds.indexes {
		for val, id := range index {
			if id == recordID {
				delete(index, val)
			}
		}
		ds.indexes[key] = index
	}
	log.Infof("Deleted record %s", recordID)
	return nil
}

// Query finds records matching a key=value condition
func (ds *YAMLDatastore) Query(key, value string) (map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make(map[string]interface{})
	if index, exists := ds.indexes[key]; exists {
		// Use index for fast lookup
		for val, id := range index {
			if fmt.Sprintf("%v", val) == value {
				if ds.txActive && ds.txData[id] == nil {
					continue // Skip deleted records in transaction
				}
				data := ds.data[id]
				if ds.txActive {
					if txData, ok := ds.txData[id]; ok && txData != nil {
						data = txData
					}
				}
				result[id] = data
			}
		}
	} else {
		// Scan all records
		for id, data := range ds.data {
			if ds.txActive && ds.txData[id] == nil {
				continue // Skip deleted records in transaction
			}
			if txData, ok := ds.txData[id]; ds.txActive && ok && txData != nil {
				data = txData
			}
			if val, ok := getNestedValue(data, key); ok && fmt.Sprintf("%v", val) == value {
				result[id] = data
			}
		}
	}
	log.Debugf("Queried %d records for %s=%s", len(result), key, value)
	return result, nil
}

// CreateIndex builds an index on a top-level key
func (ds *YAMLDatastore) CreateIndex(key string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, exists := ds.indexes[key]; exists {
		return errors.New("index already exists")
	}
	index := make(map[interface{}]string)
	for id, data := range ds.data {
		if val, ok := getNestedValue(data, key); ok {
			index[val] = id
		}
	}
	ds.indexes[key] = index
	log.Infof("Created index on %s with %d entries", key, len(index))
	return nil
}

// BeginTransaction starts a new transaction
func (ds *YAMLDatastore) BeginTransaction() (*Transaction, error) {
	ds.txMu.Lock()
	if ds.txActive {
		ds.txMu.Unlock()
		return nil, errors.New("transaction already active")
	}
	ds.txActive = true
	tx := &Transaction{
		ds:     ds,
		data:   make(map[string]interface{}),
		files:  make(map[string]bool),
		commit: false,
	}
	ds.txData = tx.data
	ds.txFiles = tx.files
	log.Info("Transaction started")
	return tx, nil
}

// Transaction represents an atomic operation
type Transaction struct {
	ds     *YAMLDatastore
	data   map[string]interface{}
	files  map[string]bool
	commit bool
}

// Set adds a record to the transaction
func (tx *Transaction) Set(recordID string, data interface{}, fileID string) error {
	tx.data[recordID] = data
	tx.files[filepath.Join(tx.ds.dir, "records_"+fileID+".yaml")] = true
	return nil
}

// Delete removes a record in the transaction
func (tx *Transaction) Delete(recordID string) error {
	tx.data[recordID] = nil
	return nil
}

// Commit applies the transaction
func (tx *Transaction) Commit() error {
	tx.ds.mu.Lock()
	defer tx.ds.mu.Unlock()
	defer tx.ds.txMu.Unlock()

	for id, data := range tx.data {
		if data == nil {
			delete(tx.ds.data, id)
		} else {
			tx.ds.data[id] = data
			tx.ds.updateIndexes(id, data)
		}
	}
	for file := range tx.files {
		tx.ds.files[file] = true
	}
	tx.ds.dirty = true
	tx.ds.invalidateCache()
	tx.ds.txActive = false
	tx.commit = true
	log.Info("Transaction committed")
	return nil
}

// Rollback discards the transaction
func (tx *Transaction) Rollback() {
	tx.ds.txMu.Unlock()
	tx.ds.mu.Lock()
	defer tx.ds.mu.Unlock()
	tx.ds.txActive = false
	tx.ds.txData = make(map[string]interface{})
	tx.ds.txFiles = make(map[string]bool)
	log.Info("Transaction rolled back")
}

// Merge combines all records
func (ds *YAMLDatastore) Merge() (map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	ds.cacheLock.RLock()
	if cached, ok := ds.cache["merged"].(map[string]interface{}); ok {
		ds.cacheLock.RUnlock()
		return cached, nil
	}
	ds.cacheLock.RUnlock()

	merged := make(map[string]interface{}, len(ds.data))
	for id, data := range ds.data {
		if ds.txActive && ds.txData[id] == nil {
			continue
		}
		if txData, ok := ds.txData[id]; ds.txActive && ok && txData != nil {
			merged[id] = txData
		} else {
			merged[id] = data
		}
	}

	ds.cacheLock.Lock()
	ds.cache["merged"] = merged
	ds.cacheLock.Unlock()
	log.Debugf("Merged %d records", len(merged))
	return merged, nil
}

// invalidateCache clears the merge cache
func (ds *YAMLDatastore) invalidateCache() {
	ds.cacheLock.Lock()
	ds.cache = make(map[string]interface{})
	ds.cacheLock.Unlock()
}

// updateIndexes updates all indexes for a record
func (ds *YAMLDatastore) updateIndexes(recordID string, data interface{}) {
	for key, index := range ds.indexes {
		for val, id := range index {
			if id == recordID {
				delete(index, val)
			}
		}
		if val, ok := getNestedValue(data, key); ok {
			index[val] = recordID
		}
		ds.indexes[key] = index
	}
}

// getNestedValue retrieves a nested value by dot-separated key
func getNestedValue(data interface{}, key string) (interface{}, bool) {
	keys := strings.Split(key, ".")
	current := data
	for _, k := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[k]
			if !ok {
				return nil, false
			}
		case map[interface{}]interface{}:
			var ok bool
			current, ok = v[k]
			if !ok {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return current, true
}

// autoSave runs periodic saves
func (ds *YAMLDatastore) autoSave() {
	interval := time.Duration(viper.GetInt("auto_save_interval")) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if err := ds.Save(); err != nil {
			log.Errorf("Auto-save error: %v", err)
		}
	}
}
