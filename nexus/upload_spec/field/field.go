// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package field

import (
	ftype "git.krd.sh/krd/woodpecker-sonatype-nexus/nexus/upload_spec/field_type"
)

// repo: https://github.com/sonatype/nexus-public.git
// file: components/nexus-repository-services/src/main/java/org/sonatype/nexus/repository/upload/UploadFieldDefinition.java
type UploadField struct {
	Name     string                `json:"name"`
	Type     ftype.UploadFieldType `json:"type,string"`
	Optional bool                  `json:"optional"`

	// optional fields
	// Group       string `json:"group,omitempty"`
	// Description string `json:"description,omitempty"`
}
