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
	"github.com/torloejborg/easykube/pkg/vars"
)

type IAddonReader interface {
	GetAddons() (map[string]IAddon, error)
	ExtractConfiguration(unconfigured IAddon) (*AddonConfig, error)
	ExtractJSON(input string) (string, bool)
	CheckAddonCompatibility() (string, error)
}

type AddonReader struct {
	EkConfig  *EasykubeConfigData
	EkContext *CobraCommandHelperImpl
}

func NewAddonReader(config IEasykubeConfig) IAddonReader {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	return &AddonReader{
		EkConfig: cfg,
	}
}

func (adr *AddonReader) CheckAddonCompatibility() (string, error) {

	// extract version from 1_easykube.js
	haystack, err := afero.ReadFile(Kube.Fs, filepath.Join(adr.EkConfig.AddonDir, constants.JS_LIB, "1-easykube.js"))
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

func (adr *AddonReader) GetAddons() (map[string]IAddon, error) {

	addons := make(map[string]IAddon)
	addonExpre, _ := regexp.Compile(`^.+\.(ek.js)$`)

	if adr.EkConfig == nil {
		panic("expected ekconfig pointer!")
	}

	if _, err := Kube.Stat(adr.EkConfig.AddonDir); err != nil {
		return nil, err
	}

	walkFunc := func(path string, entry fs.FileInfo, err error) error {
		if !entry.IsDir() && addonExpre.MatchString(entry.Name()) {
			file, openErr := Kube.Fs.Open(path)
			abs, _ := filepath.Abs(file.Name())
			if openErr != nil {
				return openErr
			}

			foundAddon := &Addon{
				Name:      entry.Name(),
				ShortName: strings.ReplaceAll(entry.Name(), ".ek.js", ""),
				File:      abs,
				RootDir:   adr.EkConfig.AddonDir,
			}

			cfg, parseErr := adr.ExtractConfiguration(foundAddon)

			if parseErr != nil {
				return errors.Join(errors.New("problem in addon: "+foundAddon.Name), parseErr)
			}

			foundAddon.Config = *cfg
			addons[foundAddon.ShortName] = foundAddon
		}

		return nil
	}

	walkErr := afero.Walk(Kube.Fs, adr.EkConfig.AddonDir, walkFunc)
	if walkErr != nil {
		return nil, walkErr
	}

	return addons, nil
}

func (adr *AddonReader) resolveExecutionOrder(
	g *Graph,
	toInstall IAddon,
	allAddons map[string]IAddon,
	out *[]IAddon, outgraph *Graph) {

	d := toInstall.GetConfig().DependsOn
	for x := range d {

		next := allAddons[d[x]]
		err := g.AddEdge(toInstall, next)

		if err != nil {
			Kube.FmtRed(err.Error())
			os.Exit(-1)
		}

		*out = append(*out, next)

		adr.resolveExecutionOrder(g, next, allAddons, out, outgraph)
		err = outgraph.AddEdge(toInstall, next)

		if err != nil {
			Kube.FmtRed(err.Error())
			os.Exit(-1)
		}
	}
}

func (adr *AddonReader) ExtractConfiguration(unconfigured IAddon) (*AddonConfig, error) {
	out := Kube.IPrinter

	code, err := afero.ReadFile(Kube.Fs, unconfigured.GetAddonFile())
	if err != nil {
		panic(err)
	}

	parsed, ok := adr.ExtractJSON(string(code))

	if len(parsed) == 0 {
		out.FmtYellow("%s Does not provide any configuration", unconfigured.GetName())
		return &AddonConfig{
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
		cfg := &AddonConfig{}
		jsonErr := json.Unmarshal([]byte(parsed), &cfg)

		// Set persistence location for all ExtraMounts
		for idx, _ := range cfg.ExtraMounts {
			cfg.ExtraMounts[idx].PersistenceDir = adr.EkConfig.PersistenceDir + "/"

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
