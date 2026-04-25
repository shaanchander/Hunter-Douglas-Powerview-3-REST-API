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
	type6GapRaw       = uint16(0x8000)
	type6Pos2Raw      = uint16(0x8000)
	type9Pos2Raw      = uint16(0x8000)
	defaultPos3Raw    = uint16(0x8000)
)

func EncodeSetPositionPacket(sequence byte, shadePct int, gapPct int, velocity int, pos2Raw uint16) []byte {
	return encodeSetPositionPacketRaw(sequence, pctToRaw(shadePct), pctToRaw(gapPct), velocity, pos2Raw)
}

func EncodeSetPositionPacketBlindOnly(sequence byte, blindPct int, velocity int, pos2Raw uint16) []byte {
	return encodeSetPositionPacketRaw(sequence, pctToRaw(blindPct), type6GapRaw, velocity, pos2Raw)
}

func encodeSetPositionPacketRaw(sequence byte, primaryRaw uint16, secondaryRaw uint16, velocity int, pos2Raw uint16) []byte {
	packet := make([]byte, 13)
	binary.LittleEndian.PutUint16(packet[0:2], setPositionOpcode)
	packet[2] = sequence
	packet[3] = payloadLength
	binary.LittleEndian.PutUint16(packet[4:6], primaryRaw)
	binary.LittleEndian.PutUint16(packet[6:8], secondaryRaw)
	binary.LittleEndian.PutUint16(packet[8:10], pos2Raw)
	binary.LittleEndian.PutUint16(packet[10:12], defaultPos3Raw)
	packet[12] = byte(velocity)
	return packet
}

func pctToRaw(pct int) uint16 {
	return uint16(pct * 100)
}

func Pos2RawForShadeType(shadeType int) (uint16, error) {
	switch shadeType {
	case 6:
		return type6Pos2Raw, nil
	case 9:
		return type9Pos2Raw, nil
	default:
		return 0, fmt.Errorf("unsupported shade type %d", shadeType)
	}
}

func PacketHexUpper(packet []byte) string {
	return strings.ToUpper(hex.EncodeToString(packet))
}
