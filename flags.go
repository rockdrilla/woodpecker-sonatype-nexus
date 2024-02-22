// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"github.com/urfave/cli/v2"
)

type Settings struct {
	RootUrl    string
	AuthPlain  string
	AuthBase64 string
	RawUploads string

	// semi-interactive mode
	Repository string
	Paths      cli.StringSlice
	Properties cli.StringSlice
}

func (p *Plugin) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "nexus.url",
			Usage:       "Sonatype Nexus URL (e.g. \"https://nexus.domain.com\")",
			EnvVars:     []string{"PLUGIN_NEXUS_URL", "NEXUS_URL"},
			Destination: &p.Settings.RootUrl,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "nexus.auth.base64",
			Usage:       "Sonatype Nexus authorization (base64-encoded, preferred)",
			EnvVars:     []string{"PLUGIN_NEXUS_AUTH_BASE64", "PLUGIN_AUTH_BASE64", "NEXUS_AUTH_BASE64"},
			Destination: &p.Settings.AuthBase64,
		},
		// https://help.sonatype.com/en/user-tokens.html#use-user-token-for-repository-authentication
		&cli.StringFlag{
			Name:        "nexus.auth.plain",
			Usage:       "Sonatype Nexus authorization (either {username}:{password} or {token name}:{token pass})",
			EnvVars:     []string{"PLUGIN_NEXUS_AUTH_PLAIN", "PLUGIN_AUTH_PLAIN", "NEXUS_AUTH_PLAIN"},
			Destination: &p.Settings.AuthPlain,
		},
		&cli.StringFlag{
			Name:        "nexus.upload",
			Usage:       "List of upload rules (JSON array)",
			EnvVars:     []string{"PLUGIN_NEXUS_UPLOAD", "PLUGIN_UPLOAD", "NEXUS_UPLOAD"},
			Destination: &p.Settings.RawUploads,
			Value:       "[]",
		},

		// semi-interactive mode
		&cli.StringFlag{
			Name:        "nexus.repository",
			Usage:       "Repository name",
			EnvVars:     []string{"NEXUS_REPOSITORY"},
			Destination: &p.Settings.Repository,
		},
		&cli.StringSliceFlag{
			Name:        "nexus.paths",
			Usage:       "Comma-separated list of paths/globs",
			EnvVars:     []string{"NEXUS_PATHS"},
			Destination: &p.Settings.Paths,
		},
		&cli.StringSliceFlag{
			Name:        "nexus.properties",
			Usage:       "Comma-separated list of properties (in form {key}={value})",
			EnvVars:     []string{"NEXUS_PROPERTIES"},
			Destination: &p.Settings.Properties,
		},
	}
}
