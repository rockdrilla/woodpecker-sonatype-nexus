// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"codeberg.org/woodpecker-plugins/go-plugin"
)

const (
	MaxAssetsPerUpload = 32
)

type Plugin struct {
	*plugin.Plugin
	Settings *Settings

	RestUrl    string
	AuthHeader string
	AuthValue  string

	UploadSpecs        map[string]UploadSpec
	UploadSpecFallback bool

	Uploads []UploadRule
}

func main() {
	p := &Plugin{
		Settings: &Settings{},
	}
	p.Plugin = plugin.New(plugin.Options{
		Name:        "woodpecker-sonatype-nexus",
		Description: "Woodpecker CI plugin to publish artifacts to Sonatype Nexus",
		Version:     "0.0.3",
		Flags:       p.Flags(),
		Execute:     p.Execute,
	})
	p.Run()
}
