package main

import (
	"os"
	"os/signal"
	"syscall"
)

// handleSignals saves state on program termination
func (ds *YAMLDatastore) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Info("Received shutdown signal, saving state...")
	if err := ds.Save(); err != nil {
		log.Errorf("Final save error: %v", err)
	}
	os.Exit(0)
}
