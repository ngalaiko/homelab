package updater

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ngalayko/dyn-dns/app/fetcher"
	"github.com/ngalayko/dyn-dns/app/provider"
)

// App Updater checks for the public ip of the server
// and updates a dns record once it changed.
type App struct {
	// public ip fetcher
	ipFetcher fetcher.Fetcher
	// dns provider api
	dnsProvider provider.Provider
	// record name to change
	record string
	// domain name to change
	domain string
	// typ is a record type
	typ provider.RecordType
	// interval between checks
	interval time.Duration
}

// New is an Updater constructor.
func New(
	p provider.Provider,
	f fetcher.Fetcher,
	domain string,
	record string,
	typ provider.RecordType,
	interval time.Duration,
) *App {
	return &App{
		dnsProvider: p,
		ipFetcher:   f,
		record:      record,
		domain:      domain,
		typ:         typ,
		interval:    interval,
	}
}

func (u *App) log(level, msg string) {
	log.Printf(
		`[%s] msg="%s" host="%s" record="%s" type="%s" interval="%s"`,
		level,
		msg,
		u.domain,
		u.record,
		u.typ,
		u.interval,
	)
}

// Run starts updater.
func (u *App) Run(ctx context.Context) error {
	u.log("INFO", "application started")

	if err := u.update(); err != nil {
		return fmt.Errorf("initial run failed: %s", err)
	}

	ticker := time.NewTicker(u.interval)
	for {
		select {
		case <-ticker.C:
			if err := u.update(); err != nil {
				u.log("ERR", err.Error())
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (u *App) update() error {
	records, err := u.dnsProvider.Get(u.domain)
	if err != nil {
		return err
	}

	currentIP, err := u.ipFetcher.Fetch()
	if err != nil {
		return err
	}

	for _, r := range records {
		fn := fullName(u.record, u.domain)

		if r.Name != u.record && r.Name != fn {
			continue
		}

		if r.Type != u.typ {
			continue
		}

		if r.Value == currentIP.String() {
			u.log("INFO", "ip is up to date")
			return nil
		}

		r.Value = currentIP.String()

		if err := u.dnsProvider.Update(r); err != nil {
			return fmt.Errorf(
				"can't update a record value to %s: %s",
				currentIP,
				err,
			)
		}

		u.log("INFO", "record updated")
		return nil
	}

	r := &provider.Record{
		Type:  provider.RecordTypeA,
		Name:  u.record,
		Value: currentIP.String(),
	}

	if err := u.dnsProvider.Create(r); err != nil {
		return fmt.Errorf(
			"can't create a record %+v: %s",
			r,
			err,
		)
	}

	u.log("INFO", "record created")
	return nil
}

func fullName(record string, domain string) string {
	switch record {
	case "@":
		return domain
	default:
		return fmt.Sprintf("%s.%s", record, domain)
	}
}
