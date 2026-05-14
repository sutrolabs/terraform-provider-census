package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/sutrolabs/terraform-provider-census/census/provider"
)

var version = "dev"

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: func() *schema.Provider { return provider.ProviderWithVersion(version) }}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/your-org/census"
	}

	plugin.Serve(opts)
}
