// Copyright 2025 Scott Friedman
// Licensed under the Apache License, Version 2.0

// Package version provides version information for the GenKit AWS project
package version

// Version is the current version of the GenKit AWS plugin
const Version = "1.0.4"

// GitCommit is the git commit hash (set during build)
var GitCommit = "unknown"

// BuildDate is the build date (set during build)
var BuildDate = "unknown"

// Info returns version information
func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_date": BuildDate,
	}
}
