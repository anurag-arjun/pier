// Package resources holds embedded bundled files.
package resources

import "embed"

//go:embed system-prompts/* prompts/*
var FS embed.FS
