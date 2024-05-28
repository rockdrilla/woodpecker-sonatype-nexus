// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rs/zerolog/log"
)

type ReadOnlyStatus struct {
	Frozen          bool   `json:"frozen"`
	SystemInitiated bool   `json:"systemInitiated"`
	Reason          string `json:"summaryReason"`
}

func (p *Plugin) GetNexusStatus(ctx context.Context) error {
	res, err := p.NexusRequest(ctx, "v1/status/writable")
	if err == nil {
		defer res.Body.Close()
		err = GenericResponseHandler(res)
	}
	if err != nil {
		log.Error().Msg("Nexus is not writable")
		return err
	}

	res, err = p.NexusRequest(ctx, "v1/read-only")
	if err == nil {
		defer res.Body.Close()
		err = GenericResponseHandler(res)
	}
	if err != nil {
		log.Error().Msg("Nexus is unable to report it's \"read-only\" status")
		return err
	}

	var roStatus ReadOnlyStatus
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&roStatus)
	if err != nil {
		log.Error().Msg("unable to decode information for \"read-only\" status")
		return err
	}

	if roStatus.Frozen {
		if roStatus.Reason == "" {
			log.Error().Msgf("Nexus is read-only (system-initiated: %v)", roStatus.SystemInitiated)
		} else {
			log.Error().Msgf("Nexus is read-only (system-initiated: %v), reason: %q", roStatus.SystemInitiated, roStatus.Reason)
		}
		return errors.New("readonly")
	}

	//TODO: determine early whether supplied credentials allows one to proceed with Sonatype Nexus

	return nil
}
