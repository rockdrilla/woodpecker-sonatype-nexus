package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type UploadSpec struct {
	Format          string        `json:"format"`
	MultipleUpload  bool          `json:"multipleUpload"`
	ComponentFields []UploadField `json:"componentFields,omitempty"`
	AssetFields     []UploadField `json:"assetFields,omitempty"`

	AllFieldNames map[string]bool
}

type UploadField struct {
	Name     string          `json:"name"`
	Type     UploadFieldType `json:"type,string"`
	Optional bool            `json:"optional"`

	// optional fields
	// Group       string `json:"group,omitempty"`
	// Description string `json:"description,omitempty"`
}

type UploadFieldType uint8

const (
	Unspecified UploadFieldType = iota
	Invalid

	File
	String
	Boolean
)

func (p *Plugin) getUploadSpecs(ctx context.Context) error {
	var rawspecs []UploadSpec

	res, err := p.NexusRequest(ctx, "v1/formats/upload-specs")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	err = GenericResponseHandler(res)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&rawspecs)
	if err != nil {
		return err
	}
	if len(rawspecs) == 0 {
		return errors.New("no upload specs were acquired from Sonatype Nexus")
	}

	p.UploadSpecs = make(map[string]UploadSpec)
	for _, s := range rawspecs {
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

		p.UploadSpecs[s.Format] = s
	}

	return nil
}

var (
	UploadFieldType_to_str map[UploadFieldType]string = map[UploadFieldType]string{
		File:    "file",
		String:  "string",
		Boolean: "boolean",
	}

	UploadFieldType_to_reflect map[UploadFieldType]reflect.Kind = map[UploadFieldType]reflect.Kind{
		File:    reflect.String,
		String:  reflect.String,
		Boolean: reflect.Bool,
	}

	UploadFieldType_from_str map[string]UploadFieldType = map[string]UploadFieldType{
		"file":    File,
		"string":  String,
		"boolean": Boolean,
	}
)

func (x UploadFieldType) IsUnspecified() bool {
	return x == Unspecified
}

func (x UploadFieldType) IsValid() bool {
	switch x {
	case File, String, Boolean:
		return true
	}
	return false
}

func (x UploadFieldType) String() string {
	s, ok := UploadFieldType_to_str[x]
	if ok {
		return s
	}
	return "UNSPECIFIED"
}

func (x UploadFieldType) ToReflectKind() reflect.Kind {
	t, ok := UploadFieldType_to_reflect[x]
	if ok {
		return t
	}
	return reflect.Invalid
}

func StringToUploadFieldType(s string) UploadFieldType {
	// if s == "" {
	// 	return Unspecified
	// }

	x, ok := UploadFieldType_from_str[strings.ToLower(s)]
	if ok {
		return x
	}
	return Invalid
}

func (x *UploadFieldType) UnmarshalJSON(b []byte) error {
	s := string(b)
	t := StringToUploadFieldType(s)
	if !t.IsUnspecified() {
		if t.IsValid() {
			*x = t
			return nil
		}

		return fmt.Errorf("incorrect type: %q", s)
	}
	return nil
}
