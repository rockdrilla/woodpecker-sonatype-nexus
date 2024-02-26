// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"github.com/rs/zerolog/log"
)

var (
	// keep map keys sorted

	fallbackUploadSpec = map[string]UploadSpec{
		"maven2": {
			MultipleUpload: true,
			ComponentFields: []UploadField{
				{
					Name: "groupId",
					Type: String,
				},
				{
					Name: "artifactId",
					Type: String,
				},
				{
					Name: "version",
					Type: String,
				},
				{
					Name:     "generate-pom",
					Type:     Boolean,
					Optional: true,
				},
				{
					Name:     "packaging",
					Type:     String,
					Optional: true,
				},
			},
			AssetFields: []UploadField{
				{
					Name: "extension",
					Type: String,
				},
				{
					Name:     "classifier",
					Type:     String,
					Optional: true,
				},
			},
		},
		"r": {
			AssetFields: []UploadField{
				{
					Name: "pathId",
					Type: String,
				},
			},
		},
		"raw": {
			MultipleUpload: true,
			ComponentFields: []UploadField{
				{
					Name: "directory",
					Type: String,
				},
			},
			AssetFields: []UploadField{
				{
					Name: "filename",
					Type: String,
				},
			},
		},
		"yum": {
			ComponentFields: []UploadField{
				{
					Name:     "directory",
					Type:     String,
					Optional: true,
				},
			},
			AssetFields: []UploadField{
				{
					Name: "filename",
					Type: String,
				},
			},
		},
	}

	// keep array values sorted

	fallbackSimpleSpecs = []string{
		"apt",
		"docker",
		"helm",
		"npm",
		"nuget",
		"pypi",
		"rubygems",
	}
)

func prepareFallbackUploadSpec() {
	for _, t := range fallbackSimpleSpecs {
		_, seen := fallbackUploadSpec[t]
		if seen {
			log.Warn().Msgf("fallback upload-spec for %q is already present!", t)
			continue
		}
		fallbackUploadSpec[t] = UploadSpec{}
	}

	for t, spec := range fallbackUploadSpec {
		spec.Format = t
	}
}
