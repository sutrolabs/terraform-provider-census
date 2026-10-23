package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/sutrolabs/terraform-provider-census/census/provider"
)

// version is set at build time via -ldflags "-X main.version=..." (see
// .goreleaser.yml and the Makefile). It is sent to Census in the User-Agent.
var version = "dev"

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: provider.New(version)}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/your-org/census"
	}

	plugin.Serve(opts)
}
