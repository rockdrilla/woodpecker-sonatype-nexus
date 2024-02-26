// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"github.com/urfave/cli/v2"
)

type Settings struct {
	RootUrl string

	AuthPlain      string
	AuthBase64     string
	AuthHttpHeader string

	RawUploads string

	// used only when "nexus.upload" is not set
	Repository string
	Paths      cli.StringSlice
	Properties cli.StringSlice
}

var (
	SensitiveEnvs []string = []string{
		"PLUGIN_NEXUS_AUTH", "PLUGIN_AUTH", "NEXUS_AUTH",
		"PLUGIN_NEXUS_AUTH_BASE64", "PLUGIN_AUTH_BASE64", "NEXUS_AUTH_BASE64",
		"PLUGIN_NEXUS_AUTH_HEADER", "PLUGIN_AUTH_HEADER", "NEXUS_AUTH_HEADER",
	}
)

func (p *Plugin) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "nexus.url",
			Usage:       "Sonatype Nexus URL (e.g. \"https://nexus.domain.com\")",
			EnvVars:     []string{"PLUGIN_NEXUS_URL", "NEXUS_URL"},
			Destination: &p.Settings.RootUrl,
			Required:    true,
		},

		// https://help.sonatype.com/en/user-tokens.html#use-user-token-for-repository-authentication
		&cli.StringFlag{
			Name:        "nexus.auth",
			Usage:       "Sonatype Nexus - HTTP Basic authorization (plain-text, either {username}:{password} or {token name}:{token pass})",
			EnvVars:     []string{"PLUGIN_NEXUS_AUTH", "PLUGIN_AUTH", "NEXUS_AUTH"},
			Destination: &p.Settings.AuthPlain,
		},
		&cli.StringFlag{
			Name:        "nexus.auth.base64",
			Usage:       "Sonatype Nexus - HTTP Basic authorization (base64-encoded, preferred)",
			EnvVars:     []string{"PLUGIN_NEXUS_AUTH_BASE64", "PLUGIN_AUTH_BASE64", "NEXUS_AUTH_BASE64"},
			Destination: &p.Settings.AuthBase64,
		},
		&cli.StringFlag{
			Name:        "nexus.auth.header",
			Usage:       "Sonatype Nexus - generic HTTP authorization header (in form {Header}={Value})",
			EnvVars:     []string{"PLUGIN_NEXUS_AUTH_HEADER", "PLUGIN_AUTH_HEADER", "NEXUS_AUTH_HEADER"},
			Destination: &p.Settings.AuthHttpHeader,
		},

		&cli.StringFlag{
			Name:        "nexus.upload",
			Usage:       "List of upload rules (JSON array)",
			EnvVars:     []string{"PLUGIN_NEXUS_UPLOAD", "PLUGIN_UPLOAD", "NEXUS_UPLOAD"},
			Destination: &p.Settings.RawUploads,
			Value:       "[]",
		},

		// used only when "nexus.upload" is not set
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
