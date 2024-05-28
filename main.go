// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"codeberg.org/woodpecker-plugins/go-plugin"

	uspec "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec"
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

	UploadSpecs        map[string]uspec.UploadSpec
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
		Version:     "0.0.2",
		Flags:       p.Flags(),
		Execute:     p.Execute,
	})
	p.Run()
}
