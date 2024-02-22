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

	p.Settings.AuthBase64 = strings.TrimSpace(p.Settings.AuthBase64)
	if (p.Settings.AuthPlain == "") && (p.Settings.AuthBase64 == "") {
		return errors.New("empty nexus.auth.plain/nexus.auth.base64")
	}
	if p.Settings.AuthBase64 == "" {
		if !strings.Contains(p.Settings.AuthPlain, ":") {
			return errors.New("nexus.auth.plain does not contain ':'")
		}
		p.Settings.AuthBase64 = base64.StdEncoding.EncodeToString([]byte(p.Settings.AuthPlain))
	}

	// paranoia area
	os.Unsetenv("PLUGIN_NEXUS_AUTH_PLAIN")
	os.Unsetenv("PLUGIN_AUTH_PLAIN")
	os.Unsetenv("PLUGIN_NEXUS_AUTH_BASE64")
	os.Unsetenv("PLUGIN_AUTH_BASE64")
	p.Settings.AuthPlain = ""

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
