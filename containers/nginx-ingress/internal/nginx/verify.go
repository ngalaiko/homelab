package nginx

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/glog"
)

// verifyClient is a client for verifying the config version.
type verifyClient struct {
	client     *http.Client
	maxRetries int
}

// newVerifyClient returns a new client pointed at the config version socket.
func newVerifyClient() *verifyClient {
	return &verifyClient{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", "/var/run/nginx-config-version.sock")
				},
			},
		},
		maxRetries: 160,
	}
}

// GetConfigVersion get version number that we put in the nginx config to verify that we're using
// the correct config.
func (c *verifyClient) GetConfigVersion() (int, error) {
	resp, err := c.client.Get("http://config-version/configVersion")
	if err != nil {
		return 0, fmt.Errorf("error getting client: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-200 response: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read the response body: %v", err)
	}
	v, err := strconv.Atoi(string(body))
	if err != nil {
		return 0, fmt.Errorf("error converting string to int: %v", err)
	}
	return v, nil
}

// WaitForCorrectVersion calls the config version endpoint until it gets the expectedVersion,
// which ensures that a new worker process has been started for that config version.
func (c *verifyClient) WaitForCorrectVersion(expectedVersion int) error {
	sleep := 25 * time.Millisecond
	for i := 1; i <= c.maxRetries; i++ {
		time.Sleep(sleep)

		version, err := c.GetConfigVersion()
		if err != nil {
			glog.V(3).Infof("Unable to fetch version: %v", err)
			continue
		}
		if version == expectedVersion {
			glog.V(3).Infof("success, version %v ensured. iterations: %v. took: %v", expectedVersion, i, time.Duration(i)*sleep)
			return nil
		}
	}
	return fmt.Errorf("could not get expected version: %v", expectedVersion)
}

const configVersionTemplateString = `server {
    listen unix:/var/run/nginx-config-version.sock;
	access_log off;
	
	{{if .OpenTracingLoadModule}}
	opentracing off;
	{{end}}

    location /configVersion {
        return 200 {{.ConfigVersion}};
    }
}
map $http_x_expected_config_version $config_version_mismatch {
	"{{.ConfigVersion}}" "";
	default "mismatch";
}`

// verifyConfigGenerator handles generating and writing the config version file.
type verifyConfigGenerator struct {
	configVersionTemplate *template.Template
}

// newVerifyConfigGenerator builds a new ConfigWriter - primarily parsing the config version template.
func newVerifyConfigGenerator() (*verifyConfigGenerator, error) {
	configVersionTemplate, err := template.New("configVersionTemplate").Parse(configVersionTemplateString)
	if err != nil {
		return nil, err
	}
	return &verifyConfigGenerator{
		configVersionTemplate: configVersionTemplate,
	}, nil
}

// GenerateVersionConfig generates the config version file.
func (c *verifyConfigGenerator) GenerateVersionConfig(configVersion int, openTracing bool) ([]byte, error) {
	var configBuffer bytes.Buffer
	templateValues := struct {
		ConfigVersion         int
		OpenTracingLoadModule bool
	}{
		configVersion,
		openTracing,
	}
	err := c.configVersionTemplate.Execute(&configBuffer, templateValues)
	if err != nil {
		return nil, err
	}

	return configBuffer.Bytes(), nil
}
