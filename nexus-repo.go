// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"
)

type NexusRepo struct {
	Name       string            `json:"name"`
	Format     string            `json:"format"`
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (p *Plugin) GetNexusRepo(ctx context.Context, repoName string) (NexusRepo, error) {
	if repoName == "" {
		log.Panic().Msg("empty repository name")
	}

	var empty NexusRepo

	res, err := p.NexusRequest(ctx, "v1/repositories/"+repoName)
	if err != nil {
		log.Error().Msgf("unable to retrieve information for repository %q", repoName)
		return empty, err
	}

	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		log.Error().Msgf("repository %q does not exist", repoName)
		return empty, errors.New("notfound")
	}

	err = GenericResponseHandler(res)
	if err != nil {
		log.Error().Msgf("unable to retrieve information for repository %q", repoName)
		return empty, err
	}

	var repo NexusRepo
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&repo)
	if err != nil {
		log.Error().Msgf("unable to decode information for repository %q", repoName)
		return empty, err
	}

	switch repo.Type {
	case "proxy", "group":
		log.Error().Msgf("repository %q is type of %q", repoName, repo.Type)
		return empty, errors.ErrUnsupported
	}

	return repo, nil
}
