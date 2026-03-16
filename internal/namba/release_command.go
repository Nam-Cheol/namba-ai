package namba

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type releaseOptions struct {
	Version string
	Bump    string
	Push    bool
	Remote  string
}

type semver struct {
	Major int
	Minor int
	Patch int
}

func (a *App) runRelease(ctx context.Context, args []string) error {
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	if !isGitRepository(root) {
		return errors.New("release requires a git repository")
	}

	opts, err := parseReleaseArgs(args)
	if err != nil {
		return err
	}

	branch, err := a.currentBranch(ctx, root)
	if err != nil {
		return fmt.Errorf("detect current branch: %w", err)
	}
	if branch != "main" {
		return fmt.Errorf("release requires the main branch, current branch is %q", branch)
	}

	status, err := a.runBinary(ctx, "git", []string{"status", "--porcelain"}, root)
	if err != nil {
		return fmt.Errorf("check working tree: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return errors.New("release requires a clean working tree")
	}

	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return err
	}
	if err := a.runValidators(ctx, root, qualityCfg); err != nil {
		return err
	}

	tagsOutput, err := a.runBinary(ctx, "git", []string{"tag", "--list", "v*"}, root)
	if err != nil {
		return fmt.Errorf("list tags: %w", err)
	}
	tags := splitLines(tagsOutput)

	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version, err = nextReleaseVersion(tags, opts.Bump)
		if err != nil {
			return err
		}
	} else {
		if _, err := parseSemver(version); err != nil {
			return err
		}
	}

	for _, tag := range tags {
		if tag == version {
			return fmt.Errorf("release tag %s already exists", version)
		}
	}

	if _, err := a.runBinary(ctx, "git", []string{"tag", version}, root); err != nil {
		return fmt.Errorf("create tag %s: %w", version, err)
	}

	fmt.Fprintf(a.stdout, "Created release tag %s\n", version)

	if !opts.Push {
		fmt.Fprintf(a.stdout, "Next: git push %s main && git push %s %s\n", opts.Remote, opts.Remote, version)
		return nil
	}

	if _, err := a.runBinary(ctx, "git", []string{"push", opts.Remote, "main"}, root); err != nil {
		return fmt.Errorf("push main: %w", err)
	}
	if _, err := a.runBinary(ctx, "git", []string{"push", opts.Remote, version}, root); err != nil {
		return fmt.Errorf("push tag %s: %w", version, err)
	}

	fmt.Fprintf(a.stdout, "Pushed main and %s to %s\n", version, opts.Remote)
	return nil
}

func parseReleaseArgs(args []string) (releaseOptions, error) {
	opts := releaseOptions{Bump: "patch", Remote: "origin"}
	bumpProvided := false

	consumeValue := func(args []string, index *int, flag string) (string, error) {
		*index = *index + 1
		if *index >= len(args) {
			return "", fmt.Errorf("%s requires a value", flag)
		}
		return args[*index], nil
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--push":
			opts.Push = true
		case "--bump":
			value, err := consumeValue(args, &i, args[i])
			if err != nil {
				return releaseOptions{}, err
			}
			opts.Bump = strings.ToLower(strings.TrimSpace(value))
			bumpProvided = true
		case "--version":
			value, err := consumeValue(args, &i, args[i])
			if err != nil {
				return releaseOptions{}, err
			}
			opts.Version = strings.TrimSpace(value)
		case "--remote":
			value, err := consumeValue(args, &i, args[i])
			if err != nil {
				return releaseOptions{}, err
			}
			opts.Remote = strings.TrimSpace(value)
		default:
			return releaseOptions{}, fmt.Errorf("unknown flag %q", args[i])
		}
	}

	if opts.Version != "" && bumpProvided {
		return releaseOptions{}, errors.New("release accepts either --version or --bump, not both")
	}
	if opts.Version == "" && !isAllowedReleaseBump(opts.Bump) {
		return releaseOptions{}, fmt.Errorf("release bump %q is not supported", opts.Bump)
	}
	if opts.Remote == "" {
		return releaseOptions{}, errors.New("release remote is required")
	}
	return opts, nil
}

func isAllowedReleaseBump(value string) bool {
	switch value {
	case "patch", "minor", "major":
		return true
	default:
		return false
	}
}

func nextReleaseVersion(tags []string, bump string) (string, error) {
	if !isAllowedReleaseBump(bump) {
		return "", fmt.Errorf("release bump %q is not supported", bump)
	}

	highest := semver{}
	found := false
	for _, tag := range tags {
		parsed, err := parseSemver(tag)
		if err != nil {
			continue
		}
		if !found || compareSemver(parsed, highest) > 0 {
			highest = parsed
			found = true
		}
	}

	next := highest
	switch bump {
	case "patch":
		if !found {
			return "v0.1.0", nil
		}
		next.Patch++
	case "minor":
		if !found {
			return "v0.1.0", nil
		}
		next.Minor++
		next.Patch = 0
	case "major":
		if !found {
			return "v1.0.0", nil
		}
		next.Major++
		next.Minor = 0
		next.Patch = 0
	}

	return formatSemver(next), nil
}

func parseSemver(tag string) (semver, error) {
	tag = strings.TrimSpace(tag)
	if !strings.HasPrefix(tag, "v") {
		return semver{}, fmt.Errorf("version %q must start with v", tag)
	}

	parts := strings.Split(strings.TrimPrefix(tag, "v"), ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("version %q must use vMAJOR.MINOR.PATCH", tag)
	}

	values := [3]int{}
	for i, part := range parts {
		if part == "" {
			return semver{}, fmt.Errorf("version %q must use vMAJOR.MINOR.PATCH", tag)
		}
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return semver{}, fmt.Errorf("version %q must use numeric MAJOR.MINOR.PATCH", tag)
		}
		values[i] = parsed
	}

	return semver{Major: values[0], Minor: values[1], Patch: values[2]}, nil
}

func compareSemver(left, right semver) int {
	if left.Major != right.Major {
		return left.Major - right.Major
	}
	if left.Minor != right.Minor {
		return left.Minor - right.Minor
	}
	return left.Patch - right.Patch
}

func formatSemver(value semver) string {
	return fmt.Sprintf("v%d.%d.%d", value.Major, value.Minor, value.Patch)
}

func splitLines(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			values = append(values, line)
		}
	}
	sort.Strings(values)
	return values
}
