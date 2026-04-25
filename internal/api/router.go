package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"hunter-douglas-powerview-3-rest-api/internal/powerview"
	"hunter-douglas-powerview-3-rest-api/internal/protocol"
)

type Config struct {
	PowerViewHost string `yaml:"POWERVIEW_HOST"`
	APIPort       string `yaml:"API_PORT"`
}

type positionRequest struct {
	SelectedShade string `json:"selectedShade"`
	ShadePct      int    `json:"shadePct"`
	GapPct        int    `json:"gapPct"`
	Velocity      *int   `json:"velocity,omitempty"`
}

type positionResponse struct {
	SentHex            string `json:"sentHex"`
	ShadePct           int    `json:"shadePct"`
	GapPct             int    `json:"gapPct"`
	DerivedBlindPct    int    `json:"derivedBlindPct"`
	UpstreamStatusCode int    `json:"upstreamStatusCode"`
}

func NewRouter(cfg Config) *gin.Engine {
	r := gin.Default()
	pvClient := powerview.NewClient(cfg.PowerViewHost)

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
		if req.ShadePct < 0 || req.ShadePct > 100 || req.GapPct < 0 || req.GapPct > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct and gapPct must be in range [0..100]"})
			return
		}
		if req.ShadePct+req.GapPct > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "shadePct + gapPct must be <= 100"})
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

		packet := protocol.EncodeSetPositionPacket(1, req.ShadePct, req.GapPct, velocity, pos2Raw)
		hexPacket := protocol.PacketHexUpper(packet)

		log.Printf("SET_POSITION hex: %s (shade=%d gap=%d velocity=%d type=%d)", hexPacket, req.ShadePct, req.GapPct, velocity, shadeType)

		statusCode, err := pvClient.SendSetPosition(req.SelectedShade, hexPacket)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send command to PowerView", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, positionResponse{
			SentHex:            hexPacket,
			ShadePct:           req.ShadePct,
			GapPct:             req.GapPct,
			DerivedBlindPct:    100 - req.ShadePct - req.GapPct,
			UpstreamStatusCode: statusCode,
		})
	})

	return r
}
