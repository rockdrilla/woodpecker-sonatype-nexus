package main

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
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
	Invariant UploadFieldType = iota
	Invalid

	File
	String
	Boolean
)

func (p *Plugin) getUploadSpecs(ctx context.Context) error {
	res, err := p.NexusRequest(ctx, "v1/formats/upload-specs")
	if err == nil {
		defer res.Body.Close()
		err = GenericResponseHandler(res)
	}
	if err != nil {
		log.Error().Msg("unable to retrieve upload specs")
		return err
	}

	var rawspecs []UploadSpec
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&rawspecs)
	if err != nil {
		log.Error().Msg("unable to decode information for upload specs")
		return err
	}
	if len(rawspecs) == 0 {
		log.Error().Msg("empty upload specs")
		return &ErrEmpty{}
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
	uploadFieldType_to_str map[UploadFieldType]string = map[UploadFieldType]string{
		Invariant: "",
		Invalid:   "INVALID",

		File:    "file",
		String:  "string",
		Boolean: "boolean",
	}

	uploadFieldType_to_reflect map[UploadFieldType]reflect.Kind = map[UploadFieldType]reflect.Kind{
		Invariant: reflect.Invalid,
		Invalid:   reflect.Invalid,

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
	return x == Invariant
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
		return Invariant
	}

	x, ok := uploadFieldType_from_str[strings.ToLower(s)]
	if ok {
		return x
	}
	return Invalid
}

func (x *UploadFieldType) UnmarshalJSON(b []byte) error {
	s := string(b)
	t := StringToUploadFieldType(s)
	if !t.IsInvariant() {
		if t.IsValid() {
			*x = t
			return nil
		}

		log.Error().Msgf("not supported UploadFieldType: %q", s)
		return errors.ErrUnsupported
	}
	return nil
}
