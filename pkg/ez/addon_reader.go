package ez

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/constants"
	"github.com/torloejborg/easykube/pkg/core"
	"github.com/torloejborg/easykube/pkg/vars"
)

type AddonReader struct {
	ek            *core.Ek
	Configuration *core.EasykubeConfigData
}

func NewAddonReader(ek *core.Ek) core.IAddonReader {

	cfg, err := ek.Config.LoadConfig()
	if err != nil {
		panic(err)
	}

	return &AddonReader{
		ek:            ek,
		Configuration: cfg,
	}
}

func (adr *AddonReader) CheckAddonCompatibility() (string, error) {

	// extract version from 1_easykube.js
	haystack, err := afero.ReadFile(adr.ek.Fs, filepath.Join(adr.Configuration.AddonDir, constants.JsLib, "1-easykube.js"))
	if err != nil {
		return "", err
	}

	if strings.Contains(vars.Version, "latest") {
		return "dev build, skipping addon catalog compatibility check", nil
	}

	vu := NewVersionUtils()
	constraint, err := vu.ExtractConstraint(string(haystack))

	if err != nil {
		return "", errors.New("semver version constraint on easykube not defined in addon repository")
	}

	ekVersion, _ := vu.ExtractVersion(vars.Version)

	if !constraint.Check(ekVersion) {
		msg := fmt.Sprintf("addon repository want easykube %s but easykube is %s\n", constraint.String(), ekVersion.String())
		return "", errors.New(msg)
	}

	msg := fmt.Sprintf("addon repository requires easykube %s easykube is %s\n", constraint.String(), ekVersion.String())
	return msg, nil
}

func (adr *AddonReader) GetAddons() (map[string]core.IAddon, error) {
	addons := make(map[string]core.IAddon)
	addonExpre := regexp.MustCompile(`^.+\.(ek.js)$`)

	if adr.Configuration == nil {
		panic("expected ekconfig pointer!")
	}

	// Resolve the root directory in case it's a symlink.
	root := adr.Configuration.AddonDir
	if _, err := adr.ek.Fs.Stat(root); err != nil {
		return nil, err
	}
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}

	visited := make(map[string]struct{})

	var walkFunc func(path string, info fs.FileInfo, err error) error

	walkFunc = func(path string, info fs.FileInfo, err error) error {

		if strings.Contains(path, ".git") {
			return nil
		}

		if err != nil {
			return err
		}

		// Resolve and guard against walking the same real path multiple times
		realPath, err := filepath.EvalSymlinks(path)
		if err == nil {
			if _, seen := visited[realPath]; seen {
				return nil
			}
			visited[realPath] = struct{}{}
		}

		// If this is a symlink to a directory, recursively walk its target.
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			// Stat target to see if it is a directory.
			tinfo, err := adr.ek.Fs.Stat(target)
			if err != nil {
				return err
			}
			if tinfo.IsDir() {
				return afero.Walk(adr.ek.Fs, target, walkFunc)
			}
			// If it’s a symlink to a file, treat it below as a regular file.
			info = tinfo
			path = target
		}

		if !info.IsDir() && addonExpre.MatchString(info.Name()) {
			file, openErr := adr.ek.Fs.Open(path)
			if openErr != nil {
				return openErr
			}
			defer file.Close()

			abs, _ := filepath.Abs(file.Name())

			foundAddon := &core.Addon{
				Name:      info.Name(),
				ShortName: strings.ReplaceAll(info.Name(), ".ek.js", ""),
				File:      abs,
				RootDir:   root,
			}

			cfg, parseErr := adr.ExtractConfiguration(foundAddon)
			if parseErr != nil {
				return errors.Join(errors.New("problem in addon: "+foundAddon.Name), parseErr)
			}
			foundAddon.Dependencies = cfg.DependsOn

			foundAddon.Config = *cfg
			addons[foundAddon.ShortName] = foundAddon
		}

		return nil
	}

	if err := afero.Walk(adr.ek.Fs, root, walkFunc); err != nil {
		return nil, err
	}

	return addons, nil
}

func (adr *AddonReader) resolveExecutionOrder(
	g *core.Graph[core.IAddon],
	toInstall core.IAddon,
	allAddons map[string]core.IAddon,
	out *[]core.IAddon, outgraph *core.Graph[core.IAddon]) {

	d := toInstall.GetConfig().DependsOn
	for x := range d {

		next := allAddons[d[x]]
		err := g.AddEdge(toInstall, next)

		if err != nil {
			adr.ek.Printer.FmtRed(err.Error())
			os.Exit(-1)
		}

		*out = append(*out, next)

		adr.resolveExecutionOrder(g, next, allAddons, out, outgraph)
		err = outgraph.AddEdge(toInstall, next)

		if err != nil {
			adr.ek.Printer.FmtRed(err.Error())
			os.Exit(-1)
		}
	}
}

func (adr *AddonReader) ExtractConfiguration(unconfigured core.IAddon) (*core.AddonConfig, error) {
	out := adr.ek.Printer

	code, err := afero.ReadFile(adr.ek.Fs, unconfigured.GetAddonFile())
	if err != nil {
		panic(err)
	}

	parsed, ok := adr.ExtractJSON(string(code))

	if len(parsed) == 0 {
		out.FmtYellow("%s Does not provide any configuration", unconfigured.GetName())
		return &core.AddonConfig{
			DependsOn:   nil,
			ExtraPorts:  nil,
			ExtraMounts: nil,
			Description: "",
		}, nil
	}

	if !ok {
		return nil, fmt.Errorf("failed to parse configuration")
	} else {

		// Parse the JSON string
		cfg := &core.AddonConfig{}
		jsonErr := json.Unmarshal([]byte(parsed), &cfg)

		// Set persistence location for all ExtraMounts
		for idx, _ := range cfg.ExtraMounts {
			cfg.ExtraMounts[idx].PersistenceDir = adr.Configuration.PersistenceDir + "/"

			// An absolute path in HostPath will be respected, and not be relative
			// to the user persistence directory
			if strings.HasPrefix(cfg.ExtraMounts[idx].HostPath, "/") {
				cfg.ExtraMounts[idx].PersistenceDir = cfg.ExtraMounts[idx].HostPath
				cfg.ExtraMounts[idx].HostPath = ""
			}
		}

		// validate port configuration
		for _, port := range cfg.ExtraPorts {
			if port.NodePort == 0 || port.HostPort == 0 {
				return nil, fmt.Errorf("%s configuration of extraPorts requires both hostPort and nodePort to be set", unconfigured.GetName())
			}
		}

		return cfg, jsonErr
	}

}

func (adr *AddonReader) ExtractJSON(input string) (string, bool) {

	// remove comments
	commentRe := regexp.MustCompile(`(?m)//.*$`)
	input = commentRe.ReplaceAllString(input, "")

	// Match assignment patterns like "let configuration =" or "configuration="
	re := regexp.MustCompile(`\b(?:let\s+)?configuration\s*=\s*`)
	loc := re.FindStringIndex(input)
	if loc == nil {
		return "", false
	}

	// Start looking for the JSON object
	startIdx := loc[1] // Position after "configuration ="
	for startIdx < len(input) && unicode.IsSpace(rune(input[startIdx])) {
		startIdx++ // Skip spaces
	}

	// Ensure we start at '{'
	if startIdx >= len(input) || input[startIdx] != '{' {
		return "", false
	}

	// Extract JSON using a stack to handle nested braces
	stack := []rune{}
	var jsonStr strings.Builder

	for i := startIdx; i < len(input); i++ {
		char := rune(input[i])
		jsonStr.WriteRune(char)

		if char == '{' {
			stack = append(stack, char) // Push to stack
		} else if char == '}' {
			stack = stack[:len(stack)-1] // Pop from stack
			if len(stack) == 0 {
				// Found the full JSON object
				return jsonStr.String(), true
			}
		}
	}

	// If we reach here, the braces were not balanced
	return "", false
}
