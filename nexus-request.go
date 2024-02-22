// SPDX-License-Identifier: Apache-2.0
// (c) 2024, Konstantin Demin

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (p *Plugin) NexusRequest(ctx context.Context, requestUrl string) (*http.Response, error) {
	return p.NexusRequestEx(ctx, http.MethodGet, requestUrl, nil, nil)
}

func (p *Plugin) NexusRequestEx(ctx context.Context, requestMethod string, requestUrl string, body io.Reader, requestSetup func(*http.Request)) (*http.Response, error) {
	c := p.HTTPClient()
	req, err := http.NewRequestWithContext(ctx, requestMethod, p.RestUrl+requestUrl, body)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.ErrUnsupported
	}
	if requestSetup != nil {
		requestSetup(req)
	}

	// TODO: support more authz schemes
	req.Header.Set("Authorization", "Basic "+p.Settings.AuthBase64)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.ErrUnsupported
	}
	return res, err
}

func GenericResponseHandler(response *http.Response) error {
	if response == nil {
		return errors.ErrUnsupported
	}

	if response.StatusCode >= 100 && response.StatusCode < 300 {
		return nil
	}

	if strings.Contains(response.Status, " ") {
		return errors.New("HTTP " + response.Status)
	}

	switch response.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("HTTP %d Unauthorized", response.StatusCode)
	case http.StatusForbidden:
		return fmt.Errorf("HTTP %d Forbidden", response.StatusCode)
	case http.StatusNotFound:
		return fmt.Errorf("HTTP %d Not found", response.StatusCode)
	}
	return fmt.Errorf("HTTP %d", response.StatusCode)
}
