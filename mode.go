// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import "os"

const (
	// DebugMode - set the framework in debug mode.
	DebugMode string = "debug"
	// ProductionMode - set the framework in production mode.
	ProductionMode string = "production"
	// TestMode - set the framework in testing mode.
	TestMode string = "test"
)

var currentMode = DebugMode

func init() {
	mode := os.Getenv("FRAMEWORK_MODE")
	if mode == "" {
		SetMode(DebugMode)
	} else {
		SetMode(mode)
	}
}

// SetMode change the current mode.
func SetMode(value string) {
	if value == DebugMode || value == ProductionMode || value == TestMode {
		currentMode = value
	} else {
		currentMode = DebugMode
	}
}

// Mode retreives the current mode.
func Mode() string {
	return currentMode
}
