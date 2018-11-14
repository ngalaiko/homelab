package main

import (
	"context"
	"flag"
	"log"
	"time"

	updater "github.com/ngalayko/dyn-dns/app"
	"github.com/ngalayko/dyn-dns/app/fetcher/ipify"
	"github.com/ngalayko/dyn-dns/app/provider"
	"github.com/ngalayko/dyn-dns/app/provider/digitalocean"
)

var (
	domain   = flag.String("domain", "example.com", "record domain to update")
	interval = flag.Duration("interval", time.Minute, "interval between checks")
	apiToken = flag.String("apiToken", "", "digitalocean api token")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	// TODO: create a config and run new instance for each record.
	if err := updater.New(
		digitalocean.New(*apiToken),
		ipify.New(),
		*domain,
		"@",
		provider.RecordTypeA,
		*interval,
	).Run(ctx); err != nil {
		log.Panicf(`[PANIC] msg="%s"`, err)
	}
}
