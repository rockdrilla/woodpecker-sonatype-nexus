// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"encoding/json"
	"fmt"
)

type ReadOnlyStatus struct {
	Frozen          bool   `json:"frozen"`
	SystemInitiated bool   `json:"systemInitiated"`
	Reason          string `json:"summaryReason"`
}

func (p *Plugin) GetNexusStatus(ctx context.Context) error {
	res, err := p.NexusRequest(ctx, "v1/status/writable")
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Nexus is not writable now (HTTP status %d)", res.StatusCode)
	}

	res, err = p.NexusRequest(ctx, "v1/read-only")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var roStatus ReadOnlyStatus
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&roStatus)
	if err != nil {
		return err
	}
	if roStatus.Frozen {
		if roStatus.Reason == "" {
			return fmt.Errorf("Nexus is read-only (system-initiated: %v)", roStatus.SystemInitiated)
		}
		return fmt.Errorf("Nexus is read-only (system-initiated: %v), reason: %q", roStatus.SystemInitiated, roStatus.Reason)
	}

	//TODO: determine early whether supplied credentials allows one to proceed with Sonatype Nexus

	return nil
}
