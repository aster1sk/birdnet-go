package config

import (
	"sync"
	"time"
)

var globalContext *Context

// OccurrenceMonitor to track species occurrences and manage state reset.
type OccurrenceMonitor struct {
	LastSpecies   string
	OccurrenceMap map[string]int
	ResetDuration time.Duration
	Mutex         sync.Mutex
	Timer         *time.Timer
}

// Context holds the overall application state, including the Settings and the OccurrenceMonitor.
type Context struct {
	Settings          *Settings
	OccurrenceMonitor *OccurrenceMonitor
}

// NewOccurrenceMonitor creates a new instance of OccurrenceMonitor with the given reset duration.
func NewOccurrenceMonitor(resetDuration time.Duration) *OccurrenceMonitor {
	return &OccurrenceMonitor{
		OccurrenceMap: make(map[string]int),
		ResetDuration: resetDuration,
	}
}

// NewContext creates a new instance of Context with the provided settings and occurrence monitor.
func NewContext(settings *Settings, occurrenceMonitor *OccurrenceMonitor) *Context {
	return &Context{
		Settings:          settings,
		OccurrenceMonitor: occurrenceMonitor,
	}
}

// TrackSpecies checks and updates the species occurrences in the OccurrenceMonitor.
func (om *OccurrenceMonitor) TrackSpecies(species string) bool {
	om.Mutex.Lock()
	defer om.Mutex.Unlock()

	if om.Timer == nil || om.LastSpecies != species {
		om.resetState(species)
		return false
	}

	om.OccurrenceMap[species]++

	if om.OccurrenceMap[species] > 1 {
		return true
	}

	return false
}

// resetState resets the state of the OccurrenceMonitor.
func (om *OccurrenceMonitor) resetState(species string) {
	om.OccurrenceMap = map[string]int{species: 1}
	om.LastSpecies = species
	if om.Timer != nil {
		om.Timer.Stop()
	}
	om.Timer = time.AfterFunc(om.ResetDuration, func() {
		om.Mutex.Lock()
		defer om.Mutex.Unlock()
		om.OccurrenceMap = make(map[string]int)
		om.Timer = nil
		om.LastSpecies = ""
	})
}

/*
	func InitGlobalContext() {
		// This function is supposed to be called after Load to make sure GlobalConfig is populated.
		globalContext = NewContext(&GlobalConfig, NewOccurrenceMonitor(10*time.Second)) // Set the duration as required
	}

func GetGlobalContext() *Context {
	if globalContext == nil {
		InitGlobalContext() // Lazy initialization, if needed
	}
	return globalContext
}
*/