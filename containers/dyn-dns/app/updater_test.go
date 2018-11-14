package updater_test

import (
	"context"
	"net"
	"testing"
	"time"

	updater "github.com/ngalayko/dyn-dns/app"
	fetcher "github.com/ngalayko/dyn-dns/app/fetcher/mock"
	"github.com/ngalayko/dyn-dns/app/provider"
	providerMock "github.com/ngalayko/dyn-dns/app/provider/mock"
)

func Test_Run__should_create_new_record(t *testing.T) {
	dnsMock := providerMock.New()
	d := "@"
	ip := net.IPv4(127, 0, 0, 1)
	u := updater.New(
		dnsMock,
		&fetcher.Mock{IP: ip},
		"example.com",
		d,
		provider.RecordTypeA,
		time.Millisecond,
	)

	ctx := context.Background()
	defer ctx.Done()

	go func() {
		if err := u.Run(ctx); err != nil {
			t.Fatalf("can't start the app: %s", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	records, err := dnsMock.Get(d)
	if err != nil {
		t.Fatalf("can't get records: %s", err)
	}

	if len(records) != 1 {
		t.Fatal("unexpected number of records")
	}

	if records[0].Name != d {
		t.Fatalf("unexpected domain name: %s", records[0].Name)
	}

	if records[0].Value != ip.String() {
		t.Fatalf("unexpected domain ip: %s", records[0].Value)
	}
}

func Test_Run__should_update_existing_record(t *testing.T) {
	dnsMock := providerMock.New()
	d := "@"
	ip := net.IPv4(127, 0, 0, 1)

	ipMock := &fetcher.Mock{IP: ip}
	u := updater.New(
		dnsMock,
		ipMock,
		"example.com",
		d,
		provider.RecordTypeA,
		time.Millisecond,
	)

	ctx := context.Background()
	defer ctx.Done()

	go func() {
		if err := u.Run(ctx); err != nil {
			t.Fatalf("can't start the app: %s", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)
	ipMock.IP = net.IPv4(1, 1, 1, 1)
	time.Sleep(10 * time.Millisecond)

	records, err := dnsMock.Get(d)
	if err != nil {
		t.Fatalf("can't get records: %s", err)
	}

	if len(records) != 1 {
		t.Fatal("unexpected number of records")
	}

	if records[0].Name != d {
		t.Fatalf("unexpected domain name: %s", records[0].Name)
	}

	if records[0].Value != ipMock.IP.String() {
		t.Fatalf("unexpected domain ip: %s", records[0].Value)
	}
}

func Test_Run__should_handle_public_ip_err(t *testing.T) {
	dnsMock := providerMock.New()
	d := "@"
	ip := net.IPv4(127, 0, 0, 1)

	ipMock := &fetcher.Mock{IP: ip}
	u := updater.New(
		dnsMock,
		ipMock,
		"example.com",
		d,
		provider.RecordTypeA,
		time.Millisecond,
	)

	ctx := context.Background()
	defer ctx.Done()

	go func() {
		if err := u.Run(ctx); err != nil {
			t.Fatalf("can't start the app: %s", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)
	var emptyIP net.IP
	ipMock.IP = emptyIP
	time.Sleep(10 * time.Millisecond)
	ipMock.IP = net.IPv4(2, 2, 2, 2)
	time.Sleep(10 * time.Millisecond)

	records, err := dnsMock.Get(d)
	if err != nil {
		t.Fatalf("can't get records: %s", err)
	}

	if len(records) != 1 {
		t.Fatal("unexpected number of records")
	}

	if records[0].Name != d {
		t.Fatalf("unexpected domain name: %s", records[0].Name)
	}

	if records[0].Value != ipMock.IP.String() {
		t.Fatalf("unexpected domain ip: %s", records[0].Value)
	}
}
