package core

type ISkaffold interface {
	CreateNewAddon(name, dest string)
}
