package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	packagerprovider "github.com/terrable-hq/terraform-provider-packager/internal/provider"
)

var version = "dev"

func main() {
	debug := flag.Bool("debug", false, "run the provider with debugger support")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		packagerprovider.New(version),
		providerserver.ServeOpts{
			Address: "registry.terraform.io/terrable-hq/packager",
			Debug:   *debug,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
