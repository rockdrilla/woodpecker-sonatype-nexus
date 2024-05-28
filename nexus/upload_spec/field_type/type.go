// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package field_type

import (
	"errors"
	"reflect"
	"strings"
)

type UploadFieldType uint8

// repo: https://github.com/sonatype/nexus-public.git
// file: components/nexus-repository-services/src/main/java/org/sonatype/nexus/repository/upload/UploadFieldDefinition.java
const (
	// internal values
	_Invariant UploadFieldType = iota
	_Invalid

	File
	String
	Boolean
)

var (
	uploadFieldType_to_str map[UploadFieldType]string = map[UploadFieldType]string{
		_Invariant: "",
		_Invalid:   "INVALID",

		File:    "file",
		String:  "string",
		Boolean: "boolean",
	}

	uploadFieldType_to_reflect map[UploadFieldType]reflect.Kind = map[UploadFieldType]reflect.Kind{
		_Invariant: reflect.Invalid,
		_Invalid:   reflect.Invalid,

		File:    reflect.String,
		String:  reflect.String,
		Boolean: reflect.Bool,
	}

	uploadFieldType_from_str map[string]UploadFieldType = map[string]UploadFieldType{
		"file":    File,
		"string":  String,
		"boolean": Boolean,
	}
)

func (x UploadFieldType) IsInvariant() bool {
	return x == _Invariant
}

func (x UploadFieldType) IsValid() bool {
	switch x {
	case File, String, Boolean:
		return true
	}
	return false
}

func (x UploadFieldType) String() string {
	s, ok := uploadFieldType_to_str[x]
	if ok {
		return s
	}
	return "INVARIANT"
}

func (x UploadFieldType) ToReflectKind() reflect.Kind {
	t, ok := uploadFieldType_to_reflect[x]
	if ok {
		return t
	}
	return reflect.Invalid
}

func StringToUploadFieldType(s string) UploadFieldType {
	if s == "" {
		return _Invariant
	}

	x, ok := uploadFieldType_from_str[strings.ToLower(s)]
	if ok {
		return x
	}
	return _Invalid
}

func (x *UploadFieldType) UnmarshalJSON(b []byte) error {
	s := string(b)
	t := StringToUploadFieldType(s)
	if !t.IsInvariant() {
		if t.IsValid() {
			*x = t
			return nil
		}
		return errors.ErrUnsupported
	}
	return nil
}
