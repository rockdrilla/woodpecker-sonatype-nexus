// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type NexusRepo struct {
	Name       string            `json:"name"`
	Format     string            `json:"format"`
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (p *Plugin) GetNexusRepo(ctx context.Context, repoName string) (NexusRepo, error) {
	var empty, repo NexusRepo

	if repoName == "" {
		return empty, errors.New("\"repoName\" parameter is empty")
	}
	res, err := p.NexusRequest(ctx, "v1/repositories/"+repoName)
	if err != nil {
		return empty, err
	}
	defer res.Body.Close()

	err = GenericResponseHandler(res)
	if err != nil {
		return empty, err
	}

	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&repo)
	if err != nil {
		return empty, err
	}

	// basic sanity check
	if repo.Type == "proxy" {
		return empty, fmt.Errorf("repository %q is type of %q", repoName, repo.Type)
	}
	// TODO: is "group" upload allowed?

	return repo, nil
}
