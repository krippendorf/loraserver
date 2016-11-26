package session

import "github.com/brocaar/lorawan"

// RXWindow defines the RX window option.
type RXWindow int8

// Available RX window options.
const (
	RX1 = iota
	RX2
)

// UplinkHistory contains meta-data of a transmission.
type UplinkHistory struct {
	FCnt         uint32
	MaxSNR       float64
	GatewayCount int
}

// NodeSession contains the informatio of a node-session (an activated node).
type NodeSession struct {
	DevAddr   lorawan.DevAddr
	AppEUI    lorawan.EUI64
	DevEUI    lorawan.EUI64
	NwkSKey   lorawan.AES128Key
	FCntUp    uint32
	FCntDown  uint32
	RelaxFCnt bool

	RXWindow    RXWindow
	RXDelay     uint8
	RX1DROffset uint8
	RX2DR       uint8

	UplinkHistory []UplinkHistory // contains the last 20 transmissions
	CFList        *lorawan.CFList
}

// AppendUplinkHistory appends an UplinkHistory item and makes sure the list
// never exceeds 20 records. In case more records are present, only the most
// recent ones will be preserved. In case of a re-transmission, the record with
// the best MaxSNR is stored.
func (b *NodeSession) AppendUplinkHistory(up UplinkHistory) {
	if count := len(b.UplinkHistory); count > 0 {
		// in case of a re-transmission, keep the record with the best MaxSNR.
		if b.UplinkHistory[count-1].FCnt == up.FCnt {
			if b.UplinkHistory[count-1].MaxSNR < up.MaxSNR {
				b.UplinkHistory = b.UplinkHistory[:count-1]
			} else {
				return
			}
		}
	}

	b.UplinkHistory = append(b.UplinkHistory, up)
	if count := len(b.UplinkHistory); count > 20 {
		b.UplinkHistory = b.UplinkHistory[count-20 : count]
	}
}

// GetPacketLossPercentage returns the percentage of packet-loss over the
// records stored in UplinkHistory.
func (b NodeSession) GetPacketLossPercentage() float64 {
	var lostPackets uint32
	var previousFCnt uint32

	for i, uh := range b.UplinkHistory {
		if i == 0 {
			previousFCnt = uh.FCnt
			continue
		}
		lostPackets += uh.FCnt - previousFCnt - 1 // there is always an expected difference of 1
		previousFCnt = uh.FCnt
	}

	return float64(lostPackets) / float64(len(b.UplinkHistory)) * 100
}