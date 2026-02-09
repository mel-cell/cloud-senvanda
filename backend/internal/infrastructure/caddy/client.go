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
func (c *Client) AddLinkDomain(projectID string, domain string, target string) error {
	routeID := fmt.Sprintf("project-%s", projectID)

	payload := map[string]interface{}{
		"@id": routeID, // ID unik agar bisa diedit/hapus nanti
		"match": []map[string]interface{}{
			{
				"host": []string{domain, domain + ":9080"},
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

	// Step 1: Hapus rute lama jika ada (berdasarkan ID)
	c.RemoveRoute(projectID)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Step 2: Tambahkan rute baru ke server default (srv0)
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

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("caddy api returned status: %d", resp.StatusCode)
	}

	return nil
}

// RemoveRoute menghapus route berdasarkan ID project
func (c *Client) RemoveRoute(projectID string) error {
	routeID := fmt.Sprintf("project-%s", projectID)
	url := fmt.Sprintf("%s/id/%s", c.BaseURL, routeID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 404 is fine, means already gone
	if resp.StatusCode != 200 && resp.StatusCode != 204 && resp.StatusCode != 404 {
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
