package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	updater "github.com/ngalayko/dyn-dns/app"
	"github.com/ngalayko/dyn-dns/app/fetcher/ipify"
	"github.com/ngalayko/dyn-dns/app/provider"
	"github.com/ngalayko/dyn-dns/app/provider/cloudflare"
	"github.com/ngalayko/dyn-dns/app/provider/digitalocean"
)

var (
	// Common flags:
	domain      = flag.String("domain", "example.com", "record domain to update")
	record      = flag.String("record", "@", "record to update")
	interval    = flag.Duration("interval", time.Minute, "interval between checks")
	apiToken    = flag.String("apiToken", "", "api token")
	dnsProvider = flag.String("dnsProvider", "", "provider type [digitalocean|cloudflare]")

	// Cloudflare flags:
	email          = flag.String("email", "", "cloudflare user account email")
	zoneIdentifier = flag.String("zoneIdentifier", "", "cloudflare zone identifier")
)

func main() {
	flag.Parse()

	app, err := createUpdater()
	if err != nil {
		log.Panicf(`[PANIC] msg="%s"`, err)
	}

	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		log.Panicf(`[PANIC] msg="%s"`, err)
	}
}

func createUpdater() (*updater.App, error) {
	var (
		dp provider.Provider
	)

	switch *dnsProvider {
	case "digitalocean":
		dp = digitalocean.New(*apiToken)
	case "cloudflare":
		dp = cloudflare.New(*apiToken, *email, *zoneIdentifier)
	default:
		return nil, fmt.Errorf("unknown dns provider: %s", *dnsProvider)
	}

	// TODO: create a config and run new instance for each record.
	return updater.New(
		dp,
		ipify.New(),
		*domain,
		*record,
		provider.RecordTypeA,
		*interval,
	), nil
}
