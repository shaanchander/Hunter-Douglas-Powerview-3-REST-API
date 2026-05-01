package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hunter-douglas-powerview-3-rest-api/internal/powerview"
	"hunter-douglas-powerview-3-rest-api/internal/protocol"
)

type Config struct {
	PowerViewHost string
	APIPort       string
}

type positionRequest struct {
	SelectedShade string `json:"selectedShade"`
	ShadePct      *int   `json:"shadePct,omitempty"`
	BlindPct      *int   `json:"blindPct"`
	Velocity      *int   `json:"velocity,omitempty"`
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

	r.GET("/v1/shades", func(c *gin.Context) {
		shades, err := pvClient.GetHomeShades()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shades from PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, shades)
	})

	r.GET("/v1/shades/:identifier", func(c *gin.Context) {
		shades, err := pvClient.GetHomeShades()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch shades from PowerView", "details": err.Error()})
			return
		}

		identifier := c.Param("identifier")

		// Try to parse as int ID first
		if id, parseErr := strconv.Atoi(identifier); parseErr == nil {
			for _, shade := range shades {
				if shade.ID == id {
					c.JSON(http.StatusOK, shade)
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "shade not found", "id": id})
			return
		}

		// Fall back to BLE name lookup
		for _, shade := range shades {
			if shade.BLEName == identifier {
				c.JSON(http.StatusOK, shade)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "shade not found", "identifier": identifier})
	})

	r.POST("/v1/position", func(c *gin.Context) {
		var req positionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if req.SelectedShade == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "selectedShade is required"})
			return
		}
		if req.BlindPct == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "blindPct is required"})
			return
		}

		shadeType, err := pvClient.GetShadeTypeByBLEName(req.SelectedShade)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to resolve shade type from PowerView", "details": err.Error()})
			return
		}

		pos2Raw, err := protocol.Pos2RawForShadeType(shadeType)
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

		switch shadeType {
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

		log.Printf("SET_POSITION hex: %s (shade=%d blindOpen=%d gap=%d velocity=%d type=%d)", hexPacket, normalizedShadePct, blindOpenPct, normalizedGapPct, velocity, shadeType)

		statusCode, err := pvClient.SendSetPosition(req.SelectedShade, hexPacket)
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
