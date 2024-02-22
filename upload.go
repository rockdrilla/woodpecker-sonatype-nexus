// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
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
	var result []UploadRule

	b := []byte(p.Settings.RawUploads)

	var base []UploadRuleBase
	err = json.Unmarshal(b, &base)
	if err != nil {
		return err
	}
	if len(base) == 0 {
		return nil
	}

	var raw []any
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}

	if len(raw) != len(base) {
		// just in case
		return fmt.Errorf("upload[] deserialization error: array length mismatch: %d != %d", len(base), len(raw))
	}

	result = make([]UploadRule, 0, len(raw))
	for i := range raw {
		if base[i].Repository == "" {
			return fmt.Errorf("upload[%d]: empty \"repository\"", i)
		}
		if len(base[i].Paths) == 0 {
			return fmt.Errorf("upload[%d]: empty \"paths\"", i)
		}

		for k, patt := range base[i].Paths {
			_, err = filepath.Glob(patt)
			if err != nil {
				return fmt.Errorf("upload[%d].paths[%d]: bad pattern %q: %v", i, k, patt, err)
			}
		}

		rtype := reflect.TypeOf(raw[i])
		if rtype.Kind() != reflect.Map {
			return fmt.Errorf("upload[%d]: not a map[string]any", i)
		}
		if rtype.Key().Kind() != reflect.String {
			return fmt.Errorf("upload[%d]: not a map[string]any", i)
		}

		m := raw[i].(map[string]any)
		ur := UploadRule{}
		ur.Repository = base[i].Repository

		ur.Paths = make([]string, len(base[i].Paths))
		copy(ur.Paths, base[i].Paths)

		for k := range m {
			switch strings.ToLower(k) {
			case "repository", "paths":
				// already handled by UploadRuleBase
				continue
			case "asset", "filename":
				// generated on per-artifact basis
				continue
			}

			rtype = reflect.TypeOf(m[k])
			switch rtype.Kind() {
			case reflect.Array,
				reflect.Chan,
				reflect.Func,
				reflect.Interface,
				reflect.Map,
				reflect.Pointer,
				reflect.Slice,
				reflect.Struct,
				reflect.UnsafePointer:
				//
				return fmt.Errorf("upload[%d]: %q is type of %q", i, k, rtype.String())
			}
			ur.Properties[k] = m[k]
		}

		result = append(result, ur)
	}

	p.Uploads = result

	return nil
}
