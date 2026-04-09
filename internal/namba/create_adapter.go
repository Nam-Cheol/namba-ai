package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const internalCreateCommandName = "__create"

type internalCreateOptions struct {
	help      bool
	operation string
	root      string
}

func (a *App) runInternalCreate(_ context.Context, args []string) error {
	options, err := parseInternalCreateArgs(args)
	if err != nil {
		return internalCreateUsageError(err)
	}
	if options.help {
		_, err := fmt.Fprint(a.stdout, internalCreateUsageText())
		return err
	}

	root, err := a.resolveInternalCreateRoot(options.root)
	if err != nil {
		return err
	}

	req, err := decodeInternalCreateRequest(a.stdin)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(a.stdout)
	encoder.SetIndent("", "  ")

	switch options.operation {
	case "preview":
		preview, err := a.previewCreate(root, req)
		if err != nil {
			return err
		}
		return encoder.Encode(preview)
	case "apply":
		result, err := a.applyCreate(root, req)
		if err != nil {
			return err
		}
		return encoder.Encode(result)
	default:
		return internalCreateUsageError(fmt.Errorf("unsupported create adapter operation %q", options.operation))
	}
}

func parseInternalCreateArgs(args []string) (internalCreateOptions, error) {
	if len(args) == 0 {
		return internalCreateOptions{}, errors.New("missing create adapter operation")
	}

	options := internalCreateOptions{
		operation: strings.TrimSpace(args[0]),
	}
	if isHelpToken(options.operation) {
		options.help = true
		return options, nil
	}

	for i := 1; i < len(args); i++ {
		switch strings.TrimSpace(args[i]) {
		case "--help", "-h":
			options.help = true
		case "--root":
			if i+1 >= len(args) {
				return internalCreateOptions{}, errors.New("missing value for --root")
			}
			i++
			options.root = strings.TrimSpace(args[i])
		default:
			return internalCreateOptions{}, fmt.Errorf("unsupported create adapter argument %q", args[i])
		}
	}

	switch options.operation {
	case "preview", "apply":
		return options, nil
	default:
		return internalCreateOptions{}, fmt.Errorf("unsupported create adapter operation %q", options.operation)
	}
}

func internalCreateUsageText() string {
	return "namba __create <preview|apply> [--root PATH]\n\nInternal create adapter for the `$namba-create` skill wrapper.\nRequests are read from stdin as JSON and results are written to stdout as JSON.\n"
}

func internalCreateUsageError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s\n\n%s", err.Error(), internalCreateUsageText())
}

func (a *App) resolveInternalCreateRoot(rootArg string) (string, error) {
	if strings.TrimSpace(rootArg) == "" {
		return a.requireProjectRoot()
	}
	root, err := filepath.Abs(rootArg)
	if err != nil {
		return "", fmt.Errorf("resolve create adapter root: %w", err)
	}
	if !exists(filepath.Join(root, nambaDir)) {
		return "", fmt.Errorf("no NambaAI project found at %s", root)
	}
	return root, nil
}

func decodeInternalCreateRequest(r io.Reader) (createRequest, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	var req createRequest
	if err := decoder.Decode(&req); err != nil {
		return createRequest{}, fmt.Errorf("decode create adapter request: %w", err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return createRequest{}, errors.New("decode create adapter request: unexpected trailing JSON input")
		}
		return createRequest{}, fmt.Errorf("decode create adapter request: %w", err)
	}
	return req, nil
}
