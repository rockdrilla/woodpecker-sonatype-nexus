// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package upload_spec

import (
	f "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec/field"
	ftype "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec/field_type"
)

var (
	// keep map keys sorted

	fallbackUploadSpec = map[string]UploadSpec{
		"maven2": {
			MultipleUpload: true,
			ComponentFields: []f.UploadField{
				{
					Name: "groupId",
					Type: ftype.String,
				},
				{
					Name: "artifactId",
					Type: ftype.String,
				},
				{
					Name: "version",
					Type: ftype.String,
				},
				{
					Name: "generate-pom",
					Type: ftype.Boolean,

					Optional: true,
				},
				{
					Name: "packaging",
					Type: ftype.String,

					Optional: true,
				},
			},
			AssetFields: []f.UploadField{
				{
					Name: "extension",
					Type: ftype.String,
				},
				{
					Name: "classifier",
					Type: ftype.String,

					Optional: true,
				},
			},
		},
		"r": {
			AssetFields: []f.UploadField{
				{
					Name: "pathId",
					Type: ftype.String,
				},
			},
		},
		"raw": {
			MultipleUpload: true,
			ComponentFields: []f.UploadField{
				{
					Name: "directory",
					Type: ftype.String,
				},
			},
			AssetFields: []f.UploadField{
				{
					Name: "filename",
					Type: ftype.String,
				},
			},
		},
		"yum": {
			ComponentFields: []f.UploadField{
				{
					Name: "directory",
					Type: ftype.String,

					Optional: true,
				},
			},
			AssetFields: []f.UploadField{
				{
					Name: "filename",
					Type: ftype.String,
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

func GetFallbackSpecs() map[string]UploadSpec {
	rv := make(map[string]UploadSpec)

	for t := range fallbackUploadSpec {
		spec := fallbackUploadSpec[t]
		spec.Format = t
		rv[t] = spec
	}

	for _, t := range fallbackSimpleSpecs {
		_, seen := rv[t]
		if seen {
			continue
		}
		rv[t] = UploadSpec{Format: t}
	}

	return rv
}
