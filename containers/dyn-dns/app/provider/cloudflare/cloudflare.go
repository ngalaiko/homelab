package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/ngalayko/dyn-dns/app/provider"
)

var (
	baseURL = "https://api.cloudflare.com/client/v4"
)

// Provider is a wrapper over Cloudflare domain API.
type Provider struct {
	apiKey         string
	email          string
	userServiceKey string
	zoneIdentifier string
}

// New is a Cloudflare provider constructor.
func New(
	apiKey string,
	email string,
	zoneIdentifier string,
) *Provider {
	if apiKey == "" {
		log.Panicf(`[PANIC] msg="%s"`, "cloudflare: api key is empty")
	}
	if email == "" {
		log.Panicf(`[PANIC] msg="%s"`, "cloudflare: email is empty")
	}
	if zoneIdentifier == "" {
		log.Panicf(`[PANIC] msg="%s"`, "cloudflare: zone identifier is empty")
	}
	return &Provider{
		apiKey:         apiKey,
		email:          email,
		zoneIdentifier: zoneIdentifier,
	}
}

// Create implements Provider interface.
func (p *Provider) Create(r *provider.Record) error {
	body, err := makeRecordData(r.Type.String(), r.Name, r.Value)
	if err != nil {
		return onError(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/zones/%s/dns_records/%s", baseURL, p.zoneIdentifier, r.ID),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return onError(err)
	}

	respData, err := p.sendRequest(req)
	if err != nil {
		return onError(err)
	}

	resp := &response{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return onError(err)
	}

	if len(resp.Errors) == 0 {
		return nil
	}

	errorMessages := make([]string, 0, len(resp.Errors))
	for _, err := range resp.Errors {
		errorMessages = append(errorMessages, err.Message)
	}

	return fmt.Errorf("cloudflare: %s", strings.Join(errorMessages, ";"))
}

// Update implements Provider interface.
func (p *Provider) Update(r *provider.Record) error {
	body, err := makeRecordData(r.Type.String(), r.Name, r.Value)
	if err != nil {
		return onError(err)
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/zones/%s/dns_records/%s", baseURL, p.zoneIdentifier, r.ID),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return onError(err)
	}

	respData, err := p.sendRequest(req)
	if err != nil {
		return onError(err)
	}

	resp := &response{}
	if err := json.Unmarshal(respData, resp); err != nil {
		return onError(err)
	}

	if len(resp.Errors) == 0 {
		return nil
	}

	errorMessages := make([]string, 0, len(resp.Errors))
	for _, err := range resp.Errors {
		errorMessages = append(errorMessages, err.Message)
	}

	return fmt.Errorf("cloudflare: %s", strings.Join(errorMessages, ";"))
}

func makeRecordData(typ, name, value string) ([]byte, error) {
	data := struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}{
		Type:    typ,
		Name:    name,
		Content: value,
	}
	return json.Marshal(data)
}

// Get implements Provider interface.
func (p *Provider) Get(domain string) ([]*provider.Record, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/zones/%s/dns_records", baseURL, p.zoneIdentifier),
		nil,
	)
	if err != nil {
		return nil, onError(err)
	}

	respData, err := p.sendRequest(req)
	if err != nil {
		return nil, onError(err)
	}

	response := &listDomainsResponse{}
	if err := json.Unmarshal(respData, response); err != nil {
		return nil, onError(err)
	}

	result := make([]*provider.Record, 0, len(response.Result))
	for _, record := range response.Result {
		result = append(result, &provider.Record{
			ID:     fmt.Sprint(record.ID),
			Type:   provider.ParseRecordType(record.Type),
			Name:   record.Name,
			Value:  record.Content,
			Domain: domain,
		})
	}

	return result, nil
}

func onError(err error) error {
	log.Printf(`[ERR] msg="cloudflare: %s"`, err)
	return fmt.Errorf("cloudflare: %s", err)
}

func (p *Provider) sendRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("X-Auth-Email", p.email)
	req.Header.Add("X-Auth-Key", p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
