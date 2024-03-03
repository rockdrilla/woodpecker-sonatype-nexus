// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package upload_spec

import (
	f "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec/field"
)

// repo: https://github.com/sonatype/nexus-public.git
// files:
// - components/nexus-repository-services/src/main/java/org/sonatype/nexus/repository/rest/api/UploadDefinitionXO.groovy
// - components/nexus-repository-services/src/main/java/org/sonatype/nexus/repository/upload/UploadDefinition.java
type UploadSpec struct {
	Format          string          `json:"format"`
	MultipleUpload  bool            `json:"multipleUpload"`
	ComponentFields []f.UploadField `json:"componentFields,omitempty"`
	AssetFields     []f.UploadField `json:"assetFields,omitempty"`

	AllFieldNames map[string]bool
}
