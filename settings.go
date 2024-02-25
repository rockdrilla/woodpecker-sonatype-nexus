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
		return errors.New("empty nexus.url")
	}
	restUrl := strings.TrimSuffix(p.Settings.RootUrl, "/") + "/service/rest/"
	_, err = url.Parse(restUrl)
	if err != nil {
		return err
	}
	p.RestUrl = restUrl

	if (p.Settings.AuthPlain == "") && (p.Settings.AuthBase64 == "") && (p.Settings.AuthHttpHeader == "") {
		return errors.New("missing nexus.auth/nexus.auth.*")
	}
	if p.Settings.AuthHttpHeader != "" {
		if !strings.Contains(p.Settings.AuthHttpHeader, "=") {
			return errors.New("nexus.auth.header does not contain '='")
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
			log.Info().Msgf("'nexus.auth.base64' is ignored while 'nexus.auth.header' is in effect")
		}
		if p.Settings.AuthPlain != "" {
			log.Info().Msgf("'nexus.auth' is ignored while 'nexus.auth.header' is in effect")
		}
	} else {
		// proceed with HTTP Basic auth
		p.AuthHeader = "Authorization"

		if p.Settings.AuthBase64 != "" {
			if p.Settings.AuthPlain != "" {
				log.Info().Msgf("'nexus.auth' is ignored while 'nexus.auth.base64' is in effect")
			}
		} else {
			if !strings.Contains(p.Settings.AuthPlain, ":") {
				return errors.New("nexus.auth does not contain ':'")
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

	// semi-interactive mode
	var ur UploadRule

	ur.Repository = p.Settings.Repository
	ur.Paths = p.Settings.Paths.Value()

	if ur.Repository == "" {
		return errors.New("empty nexus.repository")
	}
	if len(ur.Paths) == 0 {
		return errors.New("empty nexus.paths")
	}

	ur.Properties = make(map[string]any)
	for i, s := range p.Settings.Properties.Value() {
		if !strings.Contains(s, "=") {
			log.Warn().Msgf("cli.property[%d]: value does not contain '='", i)
			continue
		}

		parts := strings.SplitN(s, "=", 2)
		_, seen := ur.Properties[parts[0]]
		if seen {
			log.Warn().Msgf("cli.property[%d]: overriding previous value of %q", i, parts[0])
		}
		ur.Properties[parts[0]] = parts[1]
	}

	p.Uploads = append(p.Uploads, ur)

	return nil
}
