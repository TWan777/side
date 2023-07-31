package types

import "fmt"

type NonrefundableKey struct{}
type DisableDenomCompositionKey struct{}
type ProcessedKey struct{}

func RefundPacketKey(channelID, portID string, sequence uint64) []byte {
	return []byte(fmt.Sprintf("%s/%s/%d", channelID, portID, sequence))
}
