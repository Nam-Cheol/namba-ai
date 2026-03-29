package namba

import (
	"fmt"
	"sort"
	"strings"
)

type managedCodexMCPServerPreset struct {
	ID      string
	Command string
	Args    []string
}

var managedCodexMCPServerPresets = map[string]managedCodexMCPServerPreset{
	"context7": {
		ID:      "context7",
		Command: "npx",
		Args:    []string{"-y", "@upstash/context7-mcp"},
	},
	"playwright": {
		ID:      "playwright",
		Command: "npx",
		Args:    []string{"@playwright/mcp@latest"},
	},
	"sequential-thinking": {
		ID:      "sequential-thinking",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
	},
}

func normalizeManagedMCPServerIDs(values []string) []string {
	var normalized []string
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		id := strings.ToLower(strings.TrimSpace(value))
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		normalized = append(normalized, id)
	}
	return normalized
}

func parseCommaSeparatedValues(raw string) []string {
	return normalizeManagedMCPServerIDs(strings.Split(raw, ","))
}

func managedMCPServerPresetsForIDs(values []string) []managedCodexMCPServerPreset {
	ids := normalizeManagedMCPServerIDs(values)
	presets := make([]managedCodexMCPServerPreset, 0, len(ids))
	for _, id := range ids {
		preset, ok := managedCodexMCPServerPresets[id]
		if !ok {
			continue
		}
		presets = append(presets, preset)
	}
	return presets
}

func supportedManagedMCPServerIDs() []string {
	ids := make([]string, 0, len(managedCodexMCPServerPresets))
	for id := range managedCodexMCPServerPresets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func validateManagedMCPServerIDs(values []string) error {
	for _, id := range normalizeManagedMCPServerIDs(values) {
		if _, ok := managedCodexMCPServerPresets[id]; ok {
			continue
		}
		return fmt.Errorf("default MCP server %q is not supported (supported: %s)", id, strings.Join(supportedManagedMCPServerIDs(), ", "))
	}
	return nil
}

func formatTOMLStringArray(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%q", value))
	}
	return strings.Join(parts, ", ")
}
