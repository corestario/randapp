package types

import (
	"fmt"
	"log"

	"github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type DKGData struct {
	Data  *types.DKGData `json:"data"`
	Owner sdk.AccAddress `json:"owner"`
}

func (m DKGData) String() string {
	return fmt.Sprintf("Data: %+v, Owner: %s", m.Data, m.Owner.String())
}

// messageStore is used to store only required number of messages from every peer
type MessageStore struct {
	// Common number of messages of the same type from peers
	MessagesCount int

	// Max number of messages of the same type from one peer per round
	MaxMessagesFromPeer int

	// Map which store messages. Key is a peer's address, value is data
	Data map[string][][]byte
}

func NewMessageStore(n int) MessageStore {
	return MessageStore{
		MaxMessagesFromPeer: n,
		Data:                make(map[string][][]byte),
	}
}

func (ms *MessageStore) GetMessagesCount() int {
	return ms.MessagesCount
}

func (ms *MessageStore) GetAll() map[string][][]byte {
	return ms.Data
}

func (ms *MessageStore) Add(addr string, val []byte) {
	data := ms.Data[addr]
	if len(data) == ms.MaxMessagesFromPeer {
		log.Println("Max messages from peer!!!!!!")
		return
	}
	data = append(data, val)
	ms.Data[addr] = data
	ms.MessagesCount++
	log.Println("Message added")
}
