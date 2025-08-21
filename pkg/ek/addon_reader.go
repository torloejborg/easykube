package ek

import (
	"encoding/json"
	"fmt"
	"github.com/torloj/easykube/ekctx"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

type IAddonReader interface {
	GetAddons() map[string]*Addon
	GetExecutionOrder(wanted *Addon, addons map[string]*Addon) []*Addon
	ExtractConfiguration(unconfigured *Addon) (*AddonConfig, error)
	ExtractJSON(input string) (string, bool)
}

type AddonReader struct {
	EkConfig  *EasykubeConfigData
	EkContext *ekctx.EKContext
}

func NewAddonReader(ctx *ekctx.EKContext) IAddonReader {
	cfg, err := NewEasykubeConfig(ctx).LoadConfig()
	if err != nil {
		panic(err)
	}

	return &AddonReader{
		EkConfig:  cfg,
		EkContext: ctx,
	}
}

func (adr *AddonReader) GetAddons() map[string]*Addon {
	out := adr.EkContext.Printer
	addons := make(map[string]*Addon)
	addonExpre, _ := regexp.Compile(`^.+\.(ek.js)$`)
	if adr.EkConfig == nil {
		panic("expected ekconfig pointer!")
	}

	if _, err := os.Stat(adr.EkConfig.AddonDir); err != nil {
		return addons
	}

	walkFunc := func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() && addonExpre.MatchString(entry.Name()) {
			file, _ := os.Open(path)

			foundAddon := &Addon{
				Name:      entry.Name(),
				ShortName: strings.ReplaceAll(entry.Name(), ".ek.js", ""),
				File:      file,
				RootDir:   adr.EkConfig.AddonDir,
			}

			cfg, err := adr.ExtractConfiguration(foundAddon)

			if err != nil {
				out.FmtRed("there is issue in %s", foundAddon.Name)
				panic(err)
			}

			foundAddon.Config = *cfg
			addons[foundAddon.ShortName] = foundAddon
		}

		return nil
	}

	filepath.WalkDir(adr.EkConfig.AddonDir, walkFunc)
	return addons
}

func (adr *AddonReader) resolveExecutionOrder(
	g *Graph,
	toInstall *Addon,
	allAddons map[string]*Addon,
	out *[]*Addon, outgraph *Graph) {

	d := toInstall.Config.DependsOn
	for x := range d {

		next := allAddons[d[x]]
		err := g.AddDependency(toInstall, next)

		if err != nil {
			adr.EkContext.Printer.FmtRed(err.Error())
			os.Exit(-1)
		}

		*out = append(*out, next)

		adr.resolveExecutionOrder(g, next, allAddons, out, outgraph)
		err = outgraph.AddDependency(toInstall, next)

		if err != nil {
			adr.EkContext.Printer.FmtRed(err.Error())
			os.Exit(-1)
		}
	}
}

func (adr *AddonReader) GetExecutionOrder(wanted *Addon, addons map[string]*Addon) []*Addon {

	// transform map to list
	addonList := make([]*Addon, 0)
	for _, v := range addons {
		addonList = append(addonList, v)
	}

	// We create a Graph with the entire nodeset, this is used to detect cycles
	// when resolving the execution order
	g := NewGraph()
	g.SetNodeList(addonList)
	install := make([]*Addon, 0)
	install = append(install, wanted)

	// A new Graph will contain the result (a subset of the original Graph)
	outGraph := NewGraph()
	adr.resolveExecutionOrder(g, wanted, addons, &install, outGraph)
	outGraph.SetNodeList(install)

	order, e := outGraph.TopologicalSort()
	if e != nil {
		adr.EkContext.Printer.FmtRed(e.Error())
		os.Exit(-1)
	}

	slices.Reverse(order)
	return order
}

func (adr *AddonReader) ExtractConfiguration(unconfigured *Addon) (*AddonConfig, error) {
	out := adr.EkContext.Printer

	code, err := os.ReadFile(unconfigured.File.Name())
	if err != nil {
		panic(err)
	}

	parsed, ok := adr.ExtractJSON(string(code))

	if len(parsed) == 0 {
		out.FmtYellow("%s Does not provide any configuration", unconfigured.Name)
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
			cfg.ExtraMounts[idx].PersistenceDir = adr.EkConfig.PersistenceDir
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
