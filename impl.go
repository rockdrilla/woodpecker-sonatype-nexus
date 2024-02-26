// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func (p *Plugin) Execute(ctx context.Context) error {
	var err error

	err = p.parseSettings()
	if err != nil {
		return err
	}

	if len(p.Uploads) == 0 {
		log.Info().Msg("nothing to upload")
		return nil
	}

	err = p.GetNexusStatus(ctx)
	if err != nil {
		return err
	}

	err = p.getUploadSpecs(ctx)
	if err != nil {
		return err
	}

	repos := make(map[string]NexusRepo)
	var seen bool

	var repo NexusRepo
	var spec UploadSpec

	// validation
	for i := range p.Uploads {
		repo, seen = repos[p.Uploads[i].Repository]
		if !seen {
			repo, err = p.GetNexusRepo(ctx, p.Uploads[i].Repository)
			if err != nil {
				return fmt.Errorf("upload[%d] encountered error while getting repository information: %v", i, err)
			}

			repos[p.Uploads[i].Repository] = repo
		}

		spec, seen = p.UploadSpecs[repo.Format]
		if !seen {
			return fmt.Errorf("upload[%d] has format which is not known by upload-specs (this shouldn't happen!)", i)
		}

		del_props := make([]string, 0)
		for k := range p.Uploads[i].Properties {
			del := isInternalField(k)

			_, seen = spec.AllFieldNames[k]
			if !seen {
				del = true
			}
			if !del {
				continue
			}

			if seen {
				log.Info().Msgf("upload[%d]: %q is handled internally", i, k)
			} else {
				log.Info().Msgf("upload[%d]: %q is not used in %q spec", i, k, repo.Format)
			}
			del_props = append(del_props, k)
		}
		for _, k := range del_props {
			delete(p.Uploads[i].Properties, k)
		}
		del_props = nil

		for _, cf := range spec.ComponentFields {
			err = p.verifyUploadField(ctx, i, cf)
			if err != nil {
				return err
			}
		}

		for _, af := range spec.AssetFields {
			err = p.verifyUploadField(ctx, i, af)
			if err != nil {
				return err
			}
		}
	}

	var assets []string
	for i := range p.Uploads {
		repo = repos[p.Uploads[i].Repository]
		spec = p.UploadSpecs[repo.Format]

		// naive capacity assumption
		assets = make([]string, 0, len(p.Uploads[i].Paths))
		seen_paths := make(map[string]bool)
		for k, patt := range p.Uploads[i].Paths {
			paths, err := filepath.Glob(patt)
			if err != nil {
				// this shouldn't happen
				return err
			}
			if len(paths) == 0 {
				log.Info().Msgf("upload[%d].paths[%d]: empty match for %q", i, k, patt)
				continue
			}
			for _, path := range paths {
				_, seen := seen_paths[path]
				if seen {
					continue
				}

				err = verifyFilePath(path, fmt.Sprintf("upload[%d].paths[%d]:", i, k))
				if err != nil {
					return err
				}

				seen_paths[path] = true
				assets = append(assets, path)
			}
		}
		seen_paths = nil

		if len(assets) == 0 {
			return fmt.Errorf("upload[%d].paths[]: empty", i)
			// TODO: less strict mode?
			// log.Warn().Msgf("upload[%d]: none of paths have returned matches", i)
			// continue
		}

		if spec.MultipleUpload {
			s_end := 0
			for s_start := 0; s_start < len(assets); s_start += MaxAssetsPerUpload {
				s_end += MaxAssetsPerUpload
				if s_end > len(assets) {
					s_end = len(assets)
				}
				err = p.uploadToNexus(ctx, &p.Uploads[i], &repo, &spec, assets[s_start:s_end]...)
				if err != nil {
					return err
				}
			}
		} else {
			for _, a := range assets {
				err = p.uploadToNexus(ctx, &p.Uploads[i], &repo, &spec, a)
				if err != nil {
					return err
				}
			}
		}
	}

	log.Debug().Msg("executed")
	return nil
}

func isInternalField(fieldName string) bool {
	switch strings.ToLower(fieldName) {
	case "asset", "filename":
		return true
	}
	return false
}

func verifyFilePath(filePath, errorPrefix string) error {
	if errorPrefix == "" {
		panic(1)
	}

	if filePath == "" {
		return fmt.Errorf("%s is empty", errorPrefix)
	}

	fpath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return fmt.Errorf("%s is required but missing: %q %v", errorPrefix, filePath, err)
	}

	if !filepath.IsLocal(fpath) {
		return fmt.Errorf("%s is pointing outside of current working directory: %q", errorPrefix, filePath)
	}

	finfo, err := os.Stat(fpath)
	if err != nil {
		return fmt.Errorf("%s is required but missing: %q %v", errorPrefix, filePath, err)
	}

	if !finfo.Mode().IsRegular() {
		return fmt.Errorf("%s is required but not a regular file: %q %v", errorPrefix, filePath, err)
	}

	return nil
}

func (p *Plugin) verifyUploadField(ctx context.Context, uploadNum int, field UploadField) error {
	if isInternalField(field.Name) {
		// generated on per-artifact basis
		return nil
	}

	prop, seen := p.Uploads[uploadNum].Properties[field.Name]
	if !seen {
		if field.Optional {
			return nil
		}

		return fmt.Errorf("upload[%d]: %q is required but missing", uploadNum, field.Name)
	}

	rkind1 := reflect.TypeOf(prop).Kind()
	rkind2 := field.Type.ToReflectKind()
	if rkind1 != rkind2 {
		return fmt.Errorf("upload[%d]: %q has wrong type: %v != %v", uploadNum, field.Name, rkind1, rkind2)
	}

	switch field.Type {
	case String, File:
		s := prop.(string)
		if s == "" {
			if !field.Optional {
				return fmt.Errorf("upload[%d]: %q is required but empty", uploadNum, field.Name)
			}

			log.Info().Msgf("upload[%d]: %q is set but empty", uploadNum, field.Name)
			delete(p.Uploads[uploadNum].Properties, field.Name)
			return nil
		}

		if field.Type == String {
			// done with String
			return nil
		}

		err := verifyFilePath(s, fmt.Sprintf("upload[%d]: file %q", uploadNum, field.Name))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) uploadToNexus(ctx context.Context, upload *UploadRule, repo *NexusRepo, spec *UploadSpec, assets ...string) error {
	var err error

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	var postField string

	for _, cf := range spec.ComponentFields {
		postField = fmt.Sprintf("%s.%s", repo.Format, cf.Name)

		prop, seen := upload.Properties[cf.Name]
		if !seen {
			continue
		}

		err = writeFormFieldType(w, postField, cf.Type, prop)
		if err != nil {
			return err
		}
	}

	var assetField string
	for i, a := range assets {
		if spec.MultipleUpload {
			assetField = fmt.Sprintf("%s.asset%d", repo.Format, i+1)
		} else {
			assetField = fmt.Sprintf("%s.asset", repo.Format)
		}
		err = writeFormFile(w, assetField, a)
		if err != nil {
			return err
		}

		for _, af := range spec.AssetFields {
			switch strings.ToLower(af.Name) {
			case "asset":
				//ignored
				continue
			}

			postField = fmt.Sprintf("%s.%s", assetField, af.Name)

			switch strings.ToLower(af.Name) {
			case "filename":
				err = w.WriteField(postField, filepath.Base(a))
				if err != nil {
					return err
				}
				continue
			}

			prop, seen := upload.Properties[af.Name]
			if !seen {
				continue
			}

			err = writeFormFieldType(w, postField, af.Type, prop)
			if err != nil {
				return err
			}
		}
	}

	err = w.Close()
	if err != nil {
		return err
	}

	res, err := p.NexusRequestEx(ctx, http.MethodPost, "v1/components?repository="+upload.Repository, buf, func(r *http.Request) {
		r.Header.Set("Content-Type", w.FormDataContentType())
	})
	if err != nil {
		return err
	}

	return GenericResponseHandler(res)
}

func writeFormFile(w *multipart.Writer, fieldName string, fileName string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	part, err := w.CreateFormFile(fieldName, filepath.Base(fileName))
	if err != nil {
		return err
	}

	_, err = part.Write(data)
	return err
}

func writeFormFieldType(w *multipart.Writer, fieldName string, fieldType UploadFieldType, fieldValue any) error {
	switch fieldType {
	case File:
		return writeFormFile(w, fieldName, fieldValue.(string))
	case String:
		return w.WriteField(fieldName, fieldValue.(string))
	case Boolean:
		return w.WriteField(fieldName, strconv.FormatBool(fieldValue.(bool)))
	}

	return fmt.Errorf("unhandled field type: %v", fieldType)
}
