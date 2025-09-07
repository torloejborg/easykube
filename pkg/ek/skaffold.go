package ek

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/afero"
	"github.com/torloejborg/easykube/pkg/resources"
)

type SkaffoldImpl struct {
	AddonDir string
	Fs       afero.Fs
}

type ISkaffold interface {
	CreateNewAddon(name, dest string)
}

func NewSkaffold(addonDir string, fileFacade afero.Fs) ISkaffold {
	return &SkaffoldImpl{
		AddonDir: addonDir,
		Fs:       fileFacade,
	}
}

type model struct {
	DeploymentName string
}

func (s *SkaffoldImpl) CreateNewAddon(name, dest string) {

	err := s.Fs.MkdirAll(filepath.Join(s.AddonDir, dest, name, "manifests"), os.ModePerm)
	if err != nil {
		println("failed to create addon dir")
		log.Fatal(err)
	}

	m := model{DeploymentName: name}

	configmap := s.renderTemplate("data/skaffold/manifests/configmap.yaml", m)
	deployment := s.renderTemplate("data/skaffold/manifests/deployment.yaml", m)
	ingress := s.renderTemplate("data/skaffold/manifests/ingress.yaml", m)
	service := s.renderTemplate("data/skaffold/manifests/service.yaml", m)
	ek := s.renderTemplate("data/skaffold/ek.js", m)
	kustomization := s.renderTemplate("data/skaffold/kustomization.yaml", m)

	s.saveFile(configmap, filepath.Join(s.AddonDir, dest, name, "manifests", "configmap.yaml"))
	s.saveFile(deployment, filepath.Join(s.AddonDir, dest, name, "manifests", "deployment.yaml"))
	s.saveFile(ingress, filepath.Join(s.AddonDir, dest, name, "manifests", "ingress.yaml"))
	s.saveFile(service, filepath.Join(s.AddonDir, dest, name, "manifests", "service.yaml"))
	s.saveFile(ek, filepath.Join(s.AddonDir, dest, name, fmt.Sprintf("%s.ek.js", name)))
	s.saveFile(kustomization, filepath.Join(s.AddonDir, dest, name, "kustomization.yaml"))

}

func (a *SkaffoldImpl) renderTemplate(src string, model any) string {

	data, err := resources.AppResources.ReadFile(src)
	if err != nil {
		panic(err)
	}

	templ := template.New(src)
	templ, err = templ.Parse(string(data))
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, model)
	if err != nil {
		panic(err)
	}

	return buf.String()

}

func (a *SkaffoldImpl) saveFile(data string, dest string) {

	file, err := a.Fs.Create(dest)

	if err != nil {
		log.Fatal(err)
	}

	_, err = file.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}
}
