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

	"github.com/rs/zerolog/log"
)

func (p *Plugin) NexusRequest(ctx context.Context, requestUrl string) (*http.Response, error) {
	return p.NexusRequestEx(ctx, http.MethodGet, requestUrl, nil, nil)
}

func (p *Plugin) NexusRequestEx(ctx context.Context, requestMethod string, requestUrl string, requestBody io.Reader, requestSetup func(*http.Request)) (*http.Response, error) {
	if requestMethod == "" {
		log.Panic().Msg("empty request method")
	}
	if requestUrl == "" {
		log.Panic().Msg("empty request url")
	}

	c := p.HTTPClient()
	if c == nil {
		log.Panic().Msg("broken HTTP client")
	}

	req, err := http.NewRequestWithContext(ctx, requestMethod, p.RestUrl+requestUrl, requestBody)
	if err != nil {
		log.Error().Msgf("unable to create HTTP request: %q %q", requestMethod, "/"+requestUrl)
		return nil, err
	}
	if req == nil {
		log.Panic().Msg("nil request")
		// make analysis tools happy
		panic(1)
	}

	if requestSetup != nil {
		requestSetup(req)
	}

	req.Header.Set(p.AuthHeader, p.AuthValue)

	res, err := c.Do(req)
	if err != nil {
		log.Error().Msgf("unable to perform HTTP request: %q %q", requestMethod, "/"+requestUrl)
		return nil, err
	}
	if res == nil {
		log.Panic().Msg("nil response")
		// make analysis tools happy
		panic(1)
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		defer res.Body.Close()
		log.Error().Msgf("authentication is declined for HTTP %s %q", requestMethod, "/"+requestUrl)
		return nil, errors.New("unauthorized")
	case http.StatusForbidden:
		defer res.Body.Close()
		log.Error().Msgf("insufficient permissions for HTTP %s %q", requestMethod, "/"+requestUrl)
		return nil, errors.New("forbidden")
	}

	return res, err
}

func GenericResponseHandler(response *http.Response) error {
	if response == nil {
		log.Panic().Msg("nil response")
		// make analysis tools happy
		panic(1)
	}

	if (response.StatusCode >= http.StatusOK) && (response.StatusCode < http.StatusMultipleChoices) {
		return nil
	}

	if strings.Contains(response.Status, " ") {
		return fmt.Errorf("HTTP %s", response.Status)
	}

	// "unlikely" branch

	s := http.StatusText(response.StatusCode)
	if s != "" {
		return fmt.Errorf("HTTP %d %s", response.StatusCode, s)
	}

	return fmt.Errorf("HTTP %d Unknown return code", response.StatusCode)
}
