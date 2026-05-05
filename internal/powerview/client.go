package powerview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	host string
	http *http.Client
}

type Position struct {
	Primary   float64 `json:"primary"`
	Secondary float64 `json:"secondary"`
	Tilt      float64 `json:"tilt"`
	Velocity  int     `json:"velocity"`
}

type Shade struct {
	ID             int       `json:"id"`
	Type           int       `json:"type"`
	PTName         string    `json:"ptName"`
	BLEName        string    `json:"bleName"`
	SerialNumber   string    `json:"serialNumber"`
	SignalStrength int       `json:"signalStrength"`
	RoomID         int       `json:"roomId"`
	BatteryStatus  int       `json:"batteryStatus"`
	PowerType      int       `json:"powerType"`
	Capabilities   int       `json:"capabilities"`
	Positions      *Position `json:"positions,omitempty"`
	Firmware       *Firmware `json:"firmware,omitempty"`
}

type Firmware struct {
	Revision    int `json:"revision"`
	SubRevision int `json:"subRevision"`
	Build       int `json:"build"`
}

type Room struct {
	ID          int          `json:"id"`
	PTName      string       `json:"ptName"`
	Color       string       `json:"color"`
	Icon        string       `json:"icon"`
	Type        int          `json:"type"`
	ShadeGroups []ShadeGroup `json:"shadeGroups"`
}

type RoomDetail struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	PTName      string       `json:"ptName"`
	Color       string       `json:"color"`
	Icon        string       `json:"icon"`
	Type        int          `json:"type"`
	ShadeGroups []ShadeGroup `json:"shadeGroups"`
	Shades      []Shade      `json:"shades"`
}

type ShadeGroup struct {
	ID       int    `json:"id"`
	PTName   string `json:"ptName"`
	Order    int    `json:"order"`
	ShadeIds []int  `json:"shadeIds"`
}

type Scene struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	PTName        string `json:"ptName,omitempty"`
	NetworkNumber int    `json:"networkNumber"`
	Color         string `json:"color"`
	Icon          string `json:"icon"`
	RoomIDs       []int  `json:"roomIds"`
	ShadeIDs      []int  `json:"shadeIds"`
}

type execPayload struct {
	Hex string `json:"hex"`
}

func NewClient(host string) *Client {
	return &Client{
		host: strings.TrimRight(host, "/"),
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) GetHomeShades() ([]Shade, error) {
	endpoint := fmt.Sprintf("%s/home/shades", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/shades: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var shades []Shade
	if err := json.NewDecoder(resp.Body).Decode(&shades); err != nil {
		return nil, err
	}

	return shades, nil
}

func (c *Client) GetShadeByID(id int) (*Shade, error) {
	endpoint := fmt.Sprintf("%s/home/shades/%d", c.host, id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("shade %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/shades/%d: %s: %s", id, resp.Status, strings.TrimSpace(string(body)))
	}

	var shade Shade
	if err := json.NewDecoder(resp.Body).Decode(&shade); err != nil {
		return nil, err
	}

	return &shade, nil
}

func (c *Client) SendSetPosition(selectedShade string, hexPacket string) (int, error) {
	endpoint := fmt.Sprintf("%s/home/shades/exec?shades=%s", c.host, url.QueryEscape(selectedShade))

	payloadBytes, err := json.Marshal(execPayload{Hex: hexPacket})
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func (c *Client) GetGateway() ([]byte, error) {
	endpoint := fmt.Sprintf("%s/gateway", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from PowerView /gateway: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (c *Client) GetRooms() ([]Room, error) {
	endpoint := fmt.Sprintf("%s/home/rooms", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/rooms: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rooms []Room
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (c *Client) GetRoomByID(id int) (*RoomDetail, error) {
	endpoint := fmt.Sprintf("%s/home/rooms/%d", c.host, id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("room %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/rooms/%d: %s: %s", id, resp.Status, strings.TrimSpace(string(body)))
	}

	var roomDetail RoomDetail
	if err := json.NewDecoder(resp.Body).Decode(&roomDetail); err != nil {
		return nil, err
	}

	// Strip firmware from shades as requested
	for i := range roomDetail.Shades {
		roomDetail.Shades[i].Firmware = nil
	}

	return &roomDetail, nil
}

func (c *Client) GetSceneByID(id int) (*Scene, error) {
	endpoint := fmt.Sprintf("%s/home/scenes/%d", c.host, id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("scene %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/scenes/%d: %s: %s", id, resp.Status, strings.TrimSpace(string(body)))
	}

	var rawScene struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PTName        string `json:"ptName"`
		NetworkNumber int    `json:"networkNumber"`
		Color         string `json:"color"`
		Icon          string `json:"icon"`
		RoomIDs       []int  `json:"roomIds"`
		ShadeIDs      []int  `json:"shadeIds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawScene); err != nil {
		return nil, err
	}

	return &Scene{
		ID:            rawScene.ID,
		Name:          rawScene.PTName,
		NetworkNumber: rawScene.NetworkNumber,
		Color:         rawScene.Color,
		Icon:          rawScene.Icon,
		RoomIDs:       rawScene.RoomIDs,
		ShadeIDs:      rawScene.ShadeIDs,
	}, nil
}

func (c *Client) GetScenes() ([]Scene, error) {
	endpoint := fmt.Sprintf("%s/home/scenes", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status from PowerView /home/scenes: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rawScenes []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		PTName        string `json:"ptName"`
		NetworkNumber int    `json:"networkNumber"`
		Color         string `json:"color"`
		Icon          string `json:"icon"`
		RoomIDs       []int  `json:"roomIds"`
		ShadeIDs      []int  `json:"shadeIds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawScenes); err != nil {
		return nil, err
	}

	scenes := make([]Scene, 0, len(rawScenes))
	for _, raw := range rawScenes {
		scenes = append(scenes, Scene{
			ID:            raw.ID,
			Name:          raw.PTName,
			NetworkNumber: raw.NetworkNumber,
			Color:         raw.Color,
			Icon:          raw.Icon,
			RoomIDs:       raw.RoomIDs,
			ShadeIDs:      raw.ShadeIDs,
		})
	}

	return scenes, nil
}

func (c *Client) IsDiscoverReady() ([]byte, error) {
	endpoint := fmt.Sprintf("%s/gateway/shades/discover/ready", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from PowerView /gateway/shades/discover/ready: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (c *Client) DiscoverShades() ([]byte, error) {
	endpoint := fmt.Sprintf("%s/gateway/shades/discover/", c.host)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	// BLE scan takes longer, use a longer timeout for discovery
	longHTTP := &http.Client{Timeout: 30 * time.Second}
	resp, err := longHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from PowerView /gateway/shades/discover/: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (c *Client) TriggerScene(id int) ([]int, error) {
	// Step 1: Fetch the scene to get shadeIds and networkNumber
	scene, err := c.GetSceneByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene %d: %w", id, err)
	}

	// Step 2: Build the hex payload — F7BA0102 + networkNumber as little-endian hex (4 chars)
	// e.g., networkNumber 36466 (0x8E92) -> little-endian "928E"
	networkHex := fmt.Sprintf("%04X", scene.NetworkNumber&0xFFFF)
	reversed := fmt.Sprintf("%s%s", networkHex[2:4], networkHex[0:2])
	hexPayload := fmt.Sprintf("F7BA0102%s", reversed)

	// Step 3: For each shade in the scene, fetch the BLE name and send the command
	var statusCodes []int
	for _, shadeID := range scene.ShadeIDs {
		shade, err := c.GetShadeByID(shadeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get shade %d for scene %d: %w", shadeID, id, err)
		}

		statusCode, err := c.SendSetPosition(shade.BLEName, hexPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to send scene command to shade %s: %w", shade.BLEName, err)
		}
		statusCodes = append(statusCodes, statusCode)
	}

	return statusCodes, nil
}
