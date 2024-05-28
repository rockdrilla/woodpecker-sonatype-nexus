// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
)

type UploadRuleBase struct {
	Repository string   `json:"repository"`
	Paths      []string `json:"paths"`
}

type UploadRule struct {
	UploadRuleBase

	Properties map[string]any
}

func (p *Plugin) processRawUploads() error {
	var err error

	b := []byte(p.Settings.RawUploads)

	var base []UploadRuleBase
	err = json.Unmarshal(b, &base)
	if err != nil {
		log.Error().Msg("unable to parse upload rules")
		return err
	}
	if len(base) == 0 {
		return nil
	}

	var raw []any
	err = json.Unmarshal(b, &raw)
	if err != nil {
		log.Error().Msg("unable to parse upload rules")
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	// just in case
	b = nil

	if len(raw) != len(base) {
		// just in case
		log.Error().Msgf("upload[] deserialization error: array length mismatch: %d != %d", len(base), len(raw))
		return &ErrMalformed{}
	}

	result := make([]UploadRule, 0, len(raw))
	for i := range raw {
		if base[i].Repository == "" {
			return reportEmptySetting(fmt.Sprintf("upload[%d].repository", i))
		}
		if len(base[i].Paths) == 0 {
			return reportEmptySetting(fmt.Sprintf("upload[%d].paths", i))
		}

		for k, patt := range base[i].Paths {
			_, err = filepath.Glob(patt)
			if err != nil {
				return reportMalformedSetting(fmt.Sprintf("upload[%d].paths[%d]", i, k), fmt.Sprintf("bad pattern %q: %v", patt, err))
			}
		}

		rtype := reflect.TypeOf(raw[i])
		if rtype.Kind() != reflect.Map {
			return reportMalformedSetting(fmt.Sprintf("upload[%d]", i), fmt.Sprintf("not a map[string]any but %v", rtype.Kind()))
		}
		if rtype.Key().Kind() != reflect.String {
			return reportMalformedSetting(fmt.Sprintf("upload[%d]", i), fmt.Sprintf("not a map[string]any but map[%v]any", rtype.Key().Kind()))
		}

		m := raw[i].(map[string]any)
		ur := UploadRule{}
		ur.Repository = base[i].Repository

		ur.Paths = make([]string, len(base[i].Paths))
		copy(ur.Paths, base[i].Paths)

		for k := range m {
			switch strings.ToLower(k) {
			case "repository", "paths":
				log.Info().Msgf("upload[%d]: %q is handled by type %q", i, k, "UploadRuleBase")
				continue
			case "asset", "filename":
				log.Info().Msgf("upload[%d]: %q is handled internally on per-artifact basis", i, k)
				continue
			}

			rtype = reflect.TypeOf(m[k])
			switch rtype.Kind() {
			case reflect.Invalid,
				reflect.Array,
				reflect.Chan,
				reflect.Func,
				reflect.Interface,
				reflect.Map,
				reflect.Pointer,
				reflect.Slice,
				reflect.Struct,
				reflect.UnsafePointer:
				//
				return reportMalformedSetting(fmt.Sprintf("upload[%d]", i), fmt.Sprintf("%q is type of %q", k, rtype.String()))
			}
			ur.Properties[k] = m[k]
		}

		result = append(result, ur)
	}

	p.Uploads = result

	return nil
}
