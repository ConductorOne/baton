package uhttp

import (
	"strings"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

/*
 * Uhttp pagination handling.
 * There are three common types of pagination:
 * 1. NextLink: http header containing a url to fetch the next page
 * 2. Cursor: http body containing a token to fetch the next page
 * 3. Offset: offset + limit to fetch the next page
 *    - Subset of offset: incremental page numbers
 *
 * All of these helper functions take a bag and push the next page state on (if there is a next page).
 */

type NextLinkConfig struct {
	Header         string `json:"header,omitempty"` // HTTP header containing the next link. Defaults to "link".
	Rel            string `json:"rel,omitempty"`    // The rel value to look for in the link header. Defaults to "next".
	ResourceTypeID string `json:"resource_type_id,omitempty"`
	ResourceID     string `json:"resource_id,omitempty"`
}

// Parses the link header and returns a map of rel values to URLs.
func parseLinkHeader(header string) (map[string]string, error) {
	if header == "" {
		// Empty header is fine, it just means there are no more pages.
		return nil, nil
	}

	links := make(map[string]string)
	headerLinks := strings.Split(header, ",")
	for _, headerLink := range headerLinks {
		linkParts := strings.Split(headerLink, ";")
		if len(linkParts) < 2 {
			continue
		}
		linkUrl := strings.TrimSpace(linkParts[0])
		linkUrl = strings.Trim(linkUrl, "<>")
		var relValue string
		for _, rel := range linkParts[1:] {
			rel = strings.TrimSpace(rel)
			relParts := strings.Split(rel, "=")
			if len(relParts) < 2 {
				continue
			}
			if relParts[0] == "rel" {
				relValue = strings.Trim(relParts[1], "\"")
				break
			}
		}
		if relValue == "" {
			continue
		}
		links[relValue] = linkUrl
	}

	return links, nil
}

// WithNextLinkPagination handles nextlink pagination.
// The config is optional, and if not provided, the default config will be used.
func WithNextLinkPagination(bag *pagination.Bag, config *NextLinkConfig) DoOption {
	return func(resp *WrapperResponse) error {
		if config == nil {
			config = &NextLinkConfig{
				Header: "link",
				Rel:    "next",
			}
		}
		if config.Header == "" {
			config.Header = "link"
		}
		if config.Rel == "" {
			config.Rel = "next"
		}
		nextLinkVal := resp.Header.Get(config.Header)
		if nextLinkVal == "" {
			return nil
		}
		links, err := parseLinkHeader(nextLinkVal)
		if err != nil {
			return err
		}
		nextLink := links[config.Rel]
		if nextLink == "" {
			return nil
		}
		bag.Push(pagination.PageState{
			Token:          nextLink,
			ResourceTypeID: config.ResourceTypeID,
			ResourceID:     config.ResourceID,
		})
		return nil
	}
}
