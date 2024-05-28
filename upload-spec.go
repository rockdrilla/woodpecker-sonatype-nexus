// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	uspec "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec"
)

func (p *Plugin) getUploadSpecs(ctx context.Context) error {
	res, err := p.NexusRequest(ctx, "v1/formats/upload-specs")
	if err == nil {
		defer res.Body.Close()
		err = GenericResponseHandler(res)
	}

	for {
		if err != nil {
			p.UploadSpecFallback = true
			log.Error().Msg("unable to retrieve upload-specs")
			break
		}

		var rawspecs []uspec.UploadSpec
		dec := json.NewDecoder(res.Body)
		err = dec.Decode(&rawspecs)
		if err != nil {
			p.UploadSpecFallback = true
			log.Error().Msg("unable to decode information for upload-specs")
			break
		}

		if len(rawspecs) == 0 {
			p.UploadSpecFallback = true
			log.Error().Msg("empty upload-specs")
			break
		}

		p.UploadSpecs = make(map[string]uspec.UploadSpec)
		for _, s := range rawspecs {
			p.UploadSpecs[s.Format] = s
		}
		//lint:ignore SA4004 this is correct
		break
	}

	if p.UploadSpecFallback {
		log.Warn().Msg("using fallback upload-specs")
		p.UploadSpecs = uspec.GetFallbackSpecs()
	}

	keys := make([]string, 0, len(p.UploadSpecs))
	for k := range p.UploadSpecs {
		keys = append(keys, k)
	}

	// refill UploadSpecs
	for _, k := range keys {
		s := p.UploadSpecs[k]
		s.AllFieldNames = make(map[string]bool)
		var seen bool
		for _, f := range s.ComponentFields {
			_, seen = s.AllFieldNames[f.Name]
			if seen {
				continue
			}
			s.AllFieldNames[f.Name] = true
		}
		for _, f := range s.AssetFields {
			_, seen = s.AllFieldNames[f.Name]
			if seen {
				continue
			}
			s.AllFieldNames[f.Name] = true
		}

		p.UploadSpecs[k] = s
	}

	return nil
}
