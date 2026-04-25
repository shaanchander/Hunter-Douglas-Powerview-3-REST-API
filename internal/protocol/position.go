package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	setPositionOpcode = uint16(0x01F7)
	payloadLength     = byte(0x09)
	pos2ShadeBlindRaw = uint16(0x0000)
	pos2BlindOnlyRaw  = uint16(0x0080)
	pos3Raw           = uint16(0x0080)
)

func EncodeSetPositionPacket(sequence byte, shadePct int, gapPct int, velocity int, pos2Raw uint16) []byte {
	packet := make([]byte, 13)
	binary.LittleEndian.PutUint16(packet[0:2], setPositionOpcode)
	packet[2] = sequence
	packet[3] = payloadLength
	binary.LittleEndian.PutUint16(packet[4:6], uint16(shadePct*100))
	binary.LittleEndian.PutUint16(packet[6:8], uint16(gapPct*100))
	binary.LittleEndian.PutUint16(packet[8:10], pos2Raw)
	binary.LittleEndian.PutUint16(packet[10:12], pos3Raw)
	packet[12] = byte(velocity)
	return packet
}

func Pos2RawForShadeType(shadeType int) (uint16, error) {
	switch shadeType {
	case 6:
		return pos2BlindOnlyRaw, nil
	case 9:
		return pos2ShadeBlindRaw, nil
	default:
		return 0, fmt.Errorf("unsupported shade type %d", shadeType)
	}
}

func PacketHexUpper(packet []byte) string {
	return strings.ToUpper(hex.EncodeToString(packet))
}
