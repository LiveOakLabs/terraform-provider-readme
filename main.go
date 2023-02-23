package main

import (
	"context"
	"log"
	"terraform-provider-readme/readme"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name readme

// version sets the version of the provider.
// Override during build. Example:
// go build -ldflags "-X main.version=$VERSION"
var version string = "dev"

func main() {
	err := providerserver.Serve(context.Background(), readme.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/liveoaklabs/readme",
		// This provider requires Terraform 1.0+
		ProtocolVersion: 6,
	})
	if err != nil {
		log.Fatal("Error setting up ReadMe provider:", err)
	}
}
