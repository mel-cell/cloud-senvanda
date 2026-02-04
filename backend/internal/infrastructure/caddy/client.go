package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://caddy:2019"
	}
	return &Client{
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second, // Prevent hanging forever
		},
	}
}

// AddLinkDomain menambahkan domain baru yang mengarah ke target internal (IP:Port)
// Payload Caddy ini sedikit kompleks karena strukturnya nested.
// Kita inject route ini ke dalam http server pertama (index 0).
func (c *Client) AddLinkDomain(domain string, target string) error {
	// Construct Caddy Route JSON Payload structure
	// Ini merepresentasikan satu blok routing:
	// "match host" -> "handle reverse proxy"

	// Target format: "172.18.0.x:8080"

	routeID := fmt.Sprintf("route-%s", domain)

	payload := map[string]interface{}{
		"@id": routeID, // ID unik agar bisa diedit/hapus nanti
		"match": []map[string]interface{}{
			{
				"host": []string{domain},
			},
		},
		"handle": []map[string]interface{}{
			{
				"handler": "subroute",
				"routes": []map[string]interface{}{
					{
						"handle": []map[string]interface{}{
							{
								"handler": "reverse_proxy",
								"upstreams": []map[string]interface{}{
									{"dial": target},
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// API Endpoint untuk menambah route ke server HTTP (biasanya server "srv0")
	// Kita gunakan PUT untuk menambah/update route spesifik atau POST ke routes array.
	// Untuk keamanan, kita POST ke queue routes di server http default.
	// Path: /config/apps/http/servers/srv0/routes

	// NAMUN, karena kita pakai Caddyfile sebagai base, nama servernya mungkin random jika auto-generated,
	// atau kita perlu memastikan nama server di Caddyfile config yang dihasilkannya.
	// Biasanya server default dari Caddyfile adapter diberi nama "srv0".

	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.BaseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("caddy api returned status: %d", resp.StatusCode)
	}

	return nil
}

// Ping mengecek apakah API Caddy merespon
func (c *Client) Ping() error {
	resp, err := c.client.Get(c.BaseURL + "/config/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status: %d", resp.StatusCode)
	}
	return nil
}
