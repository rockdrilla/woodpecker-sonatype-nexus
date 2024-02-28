// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"encoding/base64"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func (p *Plugin) parseSettings() error {
	var err error

	if p.Settings.RootUrl == "" {
		return reportEmptySetting("nexus.url")
	}
	p.Settings.RootUrl = strings.TrimRight(p.Settings.RootUrl, "/")
	if p.Settings.RootUrl == "" {
		return reportMalformedSetting("nexus.url", "only slashes")
	}

	p.RestUrl = p.Settings.RootUrl + "/service/rest/"
	_, err = url.Parse(p.RestUrl)
	if err != nil {
		log.Error().Msg("unable to construct URL for REST API")
		return err
	}

	if (p.Settings.AuthPlain == "") && (p.Settings.AuthBase64 == "") && (p.Settings.AuthHttpHeader == "") {
		log.Error().Msg("missing \"nexus.auth\"/\"nexus.auth.*\"")
		return &ErrEmpty{}
	}
	if p.Settings.AuthHttpHeader != "" {
		reportSupersedingSetting("nexus.auth.header", "nexus.auth", p.Settings.AuthPlain != "")
		reportSupersedingSetting("nexus.auth.header", "nexus.auth.base64", p.Settings.AuthBase64 != "")

		if !strings.Contains(p.Settings.AuthHttpHeader, "=") {
			return reportMalformedSetting("nexus.auth.header", "does not contain '='")
		}

		parts := strings.SplitN(p.Settings.AuthHttpHeader, "=", 2)
		if parts[0] == "" {
			return reportMalformedSetting("nexus.auth.header", "empty Header")
		}
		if parts[1] == "" {
			return reportMalformedSetting("nexus.auth.header", "empty Value")
		}
		p.AuthHeader = parts[0]
		p.AuthValue = parts[1]
	} else {
		// proceed with HTTP Basic auth
		p.AuthHeader = "Authorization"

		if p.Settings.AuthBase64 != "" {
			reportSupersedingSetting("nexus.auth.base64", "nexus.auth", p.Settings.AuthPlain != "")
		} else {
			if !strings.Contains(p.Settings.AuthPlain, ":") {
				return reportMalformedSetting("nexus.auth", "does not contain ':'")
			}

			p.Settings.AuthBase64 = base64.StdEncoding.EncodeToString([]byte(p.Settings.AuthPlain))
		}

		p.AuthValue = "Basic " + p.Settings.AuthBase64
	}

	// <paranoia>
	for i := range p.Settings.flags {
		f, ok := p.Settings.flags[i].(cli.DocGenerationFlag)
		if !ok {
			continue
		}
		for _, v := range f.GetEnvVars() {
			_ = os.Unsetenv(v)
		}
	}
	p.Settings.AuthHttpHeader = ""
	p.Settings.AuthPlain = ""
	p.Settings.AuthBase64 = ""
	// </paranoia>

	err = p.processRawUploads()
	if err != nil {
		_ = reportMalformedSetting("nexus.upload", "parse error")
		return err
	}

	if len(p.Uploads) != 0 {
		reportSupersedingSetting("nexus.upload", "nexus.repository", p.Settings.Repository != "")
		reportSupersedingSetting("nexus.upload", "nexus.paths", len(p.Settings.Paths.Value()) != 0)
		reportSupersedingSetting("nexus.upload", "nexus.properties", len(p.Settings.Properties.Value()) != 0)

		return nil
	}

	log.Info().Msg("\"nexus.upload\" is empty - trying to fill it with \"inline\" parameters")

	var ur UploadRule

	ur.Repository = p.Settings.Repository
	ur.Paths = make([]string, len(p.Settings.Paths.Value()))
	copy(ur.Paths, p.Settings.Paths.Value())

	if ur.Repository == "" {
		return reportEmptySetting("nexus.repository")
	}
	if len(ur.Paths) == 0 {
		return reportEmptySetting("nexus.paths")
	}

	rawProps := p.Settings.Properties.Value()
	if len(rawProps) != 0 {
		if rawProps[0] == "" {
			return reportEmptySetting("nexus.properties")
		}
		// very naive
		for i := range rawProps {
			if rawProps[i] == "" {
				continue
			}
			switch rawProps[i][0] {
			case '{', '[':
				return reportMalformedSetting("nexus.properties", "must be plain comma-separated list, not JSON-like object")
			}
		}
	}

	ur.Properties = make(map[string]any)
	for i := range rawProps {
		if rawProps[i] == "" {
			log.Warn().Msgf("nexus.properties[%d]: empty part", i)
			continue
		}
		if !strings.Contains(rawProps[i], "=") {
			log.Warn().Msgf("nexus.properties[%d]: value does not contain '='", i)
			continue
		}

		parts := strings.SplitN(rawProps[i], "=", 2)
		_, seen := ur.Properties[parts[0]]
		if seen {
			log.Warn().Msgf("nexus.properties[%d]: overriding previous value of %q", i, parts[0])
		}
		ur.Properties[parts[0]] = parts[1]
	}

	p.Uploads = append(p.Uploads, ur)

	return nil
}

func reportEmptySetting(name string) error {
	log.Error().Msgf("\"%s\" is empty", name)
	return &ErrEmpty{}
}

func reportMalformedSetting(name, message string) error {
	log.Error().Msgf("\"%s\" is malformed: %s", name, message)
	return &ErrMalformed{}
}

func reportSupersedingSetting(settingName, supersededName string, condition bool) {
	if !condition {
		return
	}

	log.Info().Msgf("\"%s\": ignored while \"%s\" is in effect", settingName, supersededName)
}
