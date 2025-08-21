package resources

import "embed"

//go:embed data/*
var AppResources embed.FS
