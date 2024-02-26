// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func (p *Plugin) parseSettings() error {
	var err error

	p.Settings.RootUrl = strings.TrimSpace(p.Settings.RootUrl)
	if p.Settings.RootUrl == "" {
		return errors.New("nexus.url: empty")
	}
	restUrl := strings.TrimSuffix(p.Settings.RootUrl, "/") + "/service/rest/"
	_, err = url.Parse(restUrl)
	if err != nil {
		return err
	}
	p.RestUrl = restUrl

	if (p.Settings.AuthPlain == "") && (p.Settings.AuthBase64 == "") && (p.Settings.AuthHttpHeader == "") {
		return errors.New("missing 'nexus.auth'/'nexus.auth.*'")
	}
	if p.Settings.AuthHttpHeader != "" {
		if !strings.Contains(p.Settings.AuthHttpHeader, "=") {
			return errors.New("nexus.auth.header: does not contain '='")
		}

		parts := strings.SplitN(p.Settings.AuthHttpHeader, "=", 2)
		if parts[0] == "" {
			return errors.New("nexus.auth.header: empty Header")
		}
		if parts[1] == "" {
			return errors.New("nexus.auth.header: empty Value")
		}
		p.AuthHeader = parts[0]
		p.AuthValue = parts[1]

		if p.Settings.AuthBase64 != "" {
			log.Info().Msgf("nexus.auth.base64: ignored while 'nexus.auth.header' is in effect")
		}
		if p.Settings.AuthPlain != "" {
			log.Info().Msgf("nexus.auth: ignored while 'nexus.auth.header' is in effect")
		}
	} else {
		// proceed with HTTP Basic auth
		p.AuthHeader = "Authorization"

		if p.Settings.AuthBase64 != "" {
			if p.Settings.AuthPlain != "" {
				log.Info().Msgf("nexus.auth: ignored while 'nexus.auth.base64' is in effect")
			}
		} else {
			if !strings.Contains(p.Settings.AuthPlain, ":") {
				return errors.New("nexus.auth: does not contain ':'")
			}

			p.Settings.AuthBase64 = base64.StdEncoding.EncodeToString([]byte(p.Settings.AuthPlain))
		}

		p.AuthValue = "Basic " + p.Settings.AuthBase64
	}

	// <paranoia>
	for _, v := range SensitiveEnvs {
		_ = os.Unsetenv(v)
	}
	p.Settings.AuthHttpHeader = ""
	p.Settings.AuthPlain = ""
	p.Settings.AuthBase64 = ""
	// </paranoia>

	err = p.processRawUploads()
	if err != nil {
		return err
	}

	if len(p.Uploads) != 0 {
		return nil
	}

	// execution goes below only when "nexus.upload" is not set
	// e.g. semi-interactive mode
	var ur UploadRule

	ur.Repository = p.Settings.Repository
	ur.Paths = p.Settings.Paths.Value()

	if ur.Repository == "" {
		return errors.New("nexus.repository: empty")
	}
	if len(ur.Paths) == 0 {
		return errors.New("nexus.paths: empty")
	}

	rawProps := p.Settings.Properties.Value()
	if len(rawProps) == 0 {
		return errors.New("nexus.properties: empty")
	}
	if rawProps[0] == "" {
		return errors.New("nexus.properties: empty")
	}
	// very naive
	for i := range rawProps {
		if rawProps[i] == "" {
			continue
		}
		switch rawProps[i][0] {
		case '{', '[':
			return errors.New("'nexus.properties' must be plain comma-separated list, not JSON-like object")
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
