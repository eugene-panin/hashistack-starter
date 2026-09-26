package main

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Config struct {
	Vps *struct {
		ImageName string `hcl:"image_name,optional"`
	} `hcl:"vps,block"`

	Bootstrap struct {
		OpsUser       string `hcl:"ops_user"`
		PublicKeyPath string `hcl:"public_key_path"`
	} `hcl:"bootstrap,block"`
}

func loadConfig(explicit string) (Config, error) {
	cfg := Config{}
	path, err := repoPath(explicit, configFile)
	if err != nil {
		return cfg, err
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return cfg, diags
	}
	if diags := gohcl.DecodeBody(file.Body, nil, &cfg); diags.HasErrors() {
		return cfg, diags
	}
	return cfg, nil
}
