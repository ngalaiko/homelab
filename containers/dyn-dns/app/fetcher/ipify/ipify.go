package ipify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

// Fetcher is wrapper over https://www.ipify.org/ API.
type Fetcher struct{}

// New is a fetcher constructor.
func New() *Fetcher {
	return &Fetcher{}
}

type response struct {
	IP string `json:"ip"`
}

// Fetch implements Fetcher interface.
func (f *Fetcher) Fetch() (net.IP, error) {
	onError := func(err error) (net.IP, error) {
		var res net.IP
		log.Printf(`[ERR] msg="ipify: %s"`, err)
		return res, fmt.Errorf("ipify: %s", err)
	}

	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return onError(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return onError(err)
	}

	res := &response{}
	if err := json.Unmarshal(body, res); err != nil {
		return onError(err)
	}

	var ip net.IP
	if err := ip.UnmarshalText([]byte(res.IP)); err != nil {
		return onError(err)
	}

	return ip, nil
}
