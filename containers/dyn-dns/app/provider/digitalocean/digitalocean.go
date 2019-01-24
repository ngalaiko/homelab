package digitalocean

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ngalayko/dyn-dns/app/provider"
)

const (
	baseURL = "https://api.digitalocean.com"
)

// Provider is a wrapper over DigitalOcean domain API.
type Provider struct {
	apiToken string
}

// New is a DigitalOcean provider constructor.
func New(
	token string,
) *Provider {
	if token == "" {
		log.Panicf(`[PANIC] msg="%s"`, "digitalocean: api token is empty")
	}
	return &Provider{
		apiToken: token,
	}
}

// Create implements Provider interface.
func (p *Provider) Create(r *provider.Record) error {
	return fmt.Errorf("not implemented")
}

// Update implements Provider interface.
func (p *Provider) Update(r *provider.Record) error {
	body, err := updateData(r.Type.String(), r.Name, r.Value)
	if err != nil {
		return onError(err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/v2/domains/%s/records/%s", baseURL, r.Domain, r.ID),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return onError(err)
	}

	respData, err := p.sendRequest(req)
	if err != nil {
		return onError(err)
	}

	resp := &errorResponse{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return onError(err)
	}

	if len(resp.Message) != 0 {
		return onError(fmt.Errorf(resp.Message))
	}

	return nil
}

func updateData(typ, name, value string) ([]byte, error) {
	data := struct {
		Name string `json:"name"`
		Data string `json:"data"`
		Type string `json:"type"`
	}{
		Name: name,
		Data: value,
		Type: typ,
	}
	return json.Marshal(data)
}

func onError(err error) error {
	log.Printf(`[ERR] msg="digitalocean: %s"`, err)
	return fmt.Errorf("digitalocean: %s", err)
}

// Get implements Provider interface.
func (p *Provider) Get(domain string) ([]*provider.Record, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v2/domains/%s/records", baseURL, domain),
		nil,
	)
	if err != nil {
		return nil, onError(err)
	}

	respData, err := p.sendRequest(req)
	if err != nil {
		return nil, onError(err)
	}

	resp := &response{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return nil, onError(err)
	}

	result := make([]*provider.Record, 0, len(resp.Records))
	for _, d := range resp.Records {
		result = append(result, &provider.Record{
			ID:     fmt.Sprint(d.ID),
			Type:   provider.ParseRecordType(d.Type),
			Name:   d.Name,
			Value:  d.Data,
			Domain: domain,
		})
	}

	return result, nil
}

func (p *Provider) sendRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("Authorization", "Bearer "+p.apiToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
