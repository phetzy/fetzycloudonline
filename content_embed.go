// Package site embeds the content shared by the web front end and the SSH TUI.
//
// This file lives at the module root rather than under internal/ because the
// embed directive cannot reference a path outside its own package directory,
// and content.json must stay where the web build reads it. Copying it into
// the Go tree at build time would create a second source of truth, which is
// exactly what sharing the file is meant to prevent.
package site

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed content.json
var contentJSON []byte

type Link struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type Row struct {
	Label string `json:"label"`
	Body  string `json:"body"`
}

type Section struct {
	ID     string   `json:"id"`
	Tab    string   `json:"tab"`
	Label  string   `json:"label"`
	Path   string   `json:"path"`
	Meta   string   `json:"meta"`
	Title  string   `json:"title"`
	Kicker string   `json:"kicker"`
	Paras  []string `json:"paras"`
	Rows   []Row    `json:"rows"`
	Links  []Link   `json:"links"`
}

type Tab struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Header carries the copy shown in the header row that does not belong to
// any one section. Today that is just the "builds" line beside the role and
// status, which are section rows.
type Header struct {
	Builds string `json:"builds"`
}

type Content struct {
	Header   Header    `json:"header"`
	Tabs     []Tab     `json:"tabs"`
	Sections []Section `json:"sections"`
}

// Load parses the embedded content.
func Load() (Content, error) {
	var c Content
	if err := json.Unmarshal(contentJSON, &c); err != nil {
		return Content{}, fmt.Errorf("parse embedded content.json: %w", err)
	}
	return c, nil
}

// MustLoad is Load for callers that cannot proceed without content. The data is
// compiled in, so a failure here is a build-time mistake, not a runtime condition.
func MustLoad() Content {
	c, err := Load()
	if err != nil {
		panic(err)
	}
	return c
}
