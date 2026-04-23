package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
)

const (
	setPositionOpcode = uint16(0x01F7)
	payloadLength     = byte(0x09)
	unusedAxisRaw     = uint16(0x0080)
)

func EncodeSetPositionPacket(sequence byte, shadePct int, gapPct int, velocity byte) []byte {
	packet := make([]byte, 13)
	binary.LittleEndian.PutUint16(packet[0:2], setPositionOpcode)
	packet[2] = sequence
	packet[3] = payloadLength
	binary.LittleEndian.PutUint16(packet[4:6], uint16(shadePct*100))
	binary.LittleEndian.PutUint16(packet[6:8], uint16(gapPct*100))
	packet[8] = velocity
	binary.LittleEndian.PutUint16(packet[9:11], unusedAxisRaw)
	binary.LittleEndian.PutUint16(packet[11:13], unusedAxisRaw)
	return packet
}

func PacketHexUpper(packet []byte) string {
	return strings.ToUpper(hex.EncodeToString(packet))
}
