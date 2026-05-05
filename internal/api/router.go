package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"hunter-douglas-powerview-3-rest-api/internal/powerview"
	"hunter-douglas-powerview-3-rest-api/internal/protocol"
)

type Config struct {
	PowerViewHost string
	APIPort       string
}

type positionRequest struct {
	ID       int  `json:"id"`
	ShadePct *int `json:"shadePct,omitempty"`
	BlindPct *int `json:"blindPct"`
	Velocity *int `json:"velocity,omitempty"`
}

type positionResponse struct {
	SentHex            string `json:"sentHex"`
	ShadePct           int    `json:"shadePct"`
	BlindPct           int    `json:"blindPct"`
	UpstreamStatusCode int    `json:"upstreamStatusCode"`
}

func NewRouter(cfg Config) *gin.Engine {
	r := gin.Default()
	pvClient := powerview.NewClient(cfg.PowerViewHost)

	r.GET("/v1/gateway", func(c *gin.Context) {
		body, err := pvClient.GetGateway()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch gateway from PowerView", "details": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
	})

	r.GET("/v1/gateway/flash", func(c *gin.Context) {
		body, err := pvClient.GatewayFlash()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to flash gateway LED from PowerView", "details": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
	})

	r.GET("/v1/shades", func(c *gin.Context) {
		shades, err := pvClient.GetHomeShades()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shades from PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, shades)
	})

	r.GET("/v1/shades/:id", func(c *gin.Context) {
		shades, err := pvClient.GetHomeShades()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shades from PowerView", "details": err.Error()})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
			return
		}

		for _, shade := range shades {
			if shade.ID == id {
				c.JSON(http.StatusOK, shade)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "shade not found", "id": id})
	})

	r.GET("/v1/shades/:id/jog", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
			return
		}

		shade, err := pvClient.GetShadeByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "shade not found", "id": id})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shade from PowerView", "details": err.Error()})
			return
		}

		const jogHex = "F7110A0103"
		statusCode, err := pvClient.SendSetPosition(shade.BLEName, jogHex)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send jog command to PowerView", "details": err.Error()})
			return
		}

		log.Printf("JOG shade id=%d ble=%s hex=%s status=%d", id, shade.BLEName, jogHex, statusCode)

		c.JSON(http.StatusOK, gin.H{"shadeId": id, "sentHex": jogHex, "upstreamStatusCode": statusCode})
	})

	r.GET("/v1/rooms", func(c *gin.Context) {
		rooms, err := pvClient.GetRooms()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch rooms from PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rooms)
	})

	r.GET("/v1/rooms/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
			return
		}

		room, err := pvClient.GetRoomByID(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "room not found", "id": id})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch room from PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, room)
	})

	r.GET("/v1/scenes", func(c *gin.Context) {
		scenes, err := pvClient.GetScenes()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch scenes from PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, scenes)
	})

	r.GET("/v1/scenes/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
			return
		}

		body, err := pvClient.GetSceneByIDRaw(id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.JSON(http.StatusNotFound, gin.H{"error": "scene not found", "id": id})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch scene from PowerView", "details": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
	})

	r.GET("/v1/scenes/:id/trigger", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a valid integer"})
			return
		}

		statusCodes, err := pvClient.TriggerScene(id)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to trigger scene on PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"sceneId": id, "upstreamStatusCodes": statusCodes})
	})

	r.GET("/v1/discover/ready", func(c *gin.Context) {
		body, err := pvClient.IsDiscoverReady()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to check discover readiness from PowerView", "details": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
	})

	r.GET("/v1/discover", func(c *gin.Context) {
		readyBody, err := pvClient.IsDiscoverReady()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to check discover readiness from PowerView", "details": err.Error()})
			return
		}

		var readyResp struct {
			Ready bool `json:"ready"`
		}
		if err := json.Unmarshal(readyBody, &readyResp); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to parse discover readiness response", "details": err.Error()})
			return
		}
		if !readyResp.Ready {
			c.JSON(http.StatusConflict, gin.H{"error": "gateway is not ready for discovery", "details": "a discovery scan is already in progress"})
			return
		}

		body, err := pvClient.DiscoverShades()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to discover shades from PowerView", "details": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", body)
	})

	r.POST("/v1/position", func(c *gin.Context) {
		var req positionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		if req.BlindPct == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "blindPct is required"})
			return
		}

		shade, err := pvClient.GetShadeByID(req.ID)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shade from PowerView", "details": err.Error()})
			return
		}

		pos2Raw, err := protocol.Pos2RawForShadeType(shade.Type)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported shade type", "details": err.Error()})
			return
		}

		velocity := 0
		if req.Velocity != nil {
			if *req.Velocity < 10 || *req.Velocity > 255 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "velocity must be in range [10..255] when provided"})
				return
			}
			velocity = *req.Velocity
		}

		blindOpenPct := *req.BlindPct
		normalizedShadePct := 0
		normalizedGapPct := 0
		responseShadePct := 0
		useBlindOnlyEncoding := false

		switch shade.Type {
		case 9:
			if req.ShadePct == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct is required for shade+blind devices"})
				return
			}
			if *req.ShadePct < 0 || *req.ShadePct > 100 || blindOpenPct < 0 || blindOpenPct > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct and blindPct must be in range [0..100]"})
				return
			}
			if *req.ShadePct+blindOpenPct > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct + blindPct must be <= 100"})
				return
			}

			normalizedShadePct = *req.ShadePct
			normalizedGapPct = 100 - *req.ShadePct - blindOpenPct
			responseShadePct = *req.ShadePct
		case 6:
			if req.ShadePct != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct must be omitted for blind-only devices"})
				return
			}
			if blindOpenPct < 0 || blindOpenPct > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "blindPct must be in range [0..100]"})
				return
			}

			normalizedShadePct = 100 - blindOpenPct
			normalizedGapPct = -1
			responseShadePct = -1
			useBlindOnlyEncoding = true
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported shade type", "details": "unsupported shade type"})
			return
		}

		var packet []byte
		if useBlindOnlyEncoding {
			packet = protocol.EncodeSetPositionPacketBlindOnly(1, normalizedShadePct, velocity, pos2Raw)
		} else {
			packet = protocol.EncodeSetPositionPacket(1, normalizedShadePct, normalizedGapPct, velocity, pos2Raw)
		}
		hexPacket := protocol.PacketHexUpper(packet)

		log.Printf("SET_POSITION hex: %s (shade=%d blindOpen=%d gap=%d velocity=%d type=%d)", hexPacket, normalizedShadePct, blindOpenPct, normalizedGapPct, velocity, shade.Type)

		statusCode, err := pvClient.SendSetPosition(shade.BLEName, hexPacket)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send command to PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, positionResponse{
			SentHex:            hexPacket,
			ShadePct:           responseShadePct,
			BlindPct:           blindOpenPct,
			UpstreamStatusCode: statusCode,
		})
	})

	return r
}
