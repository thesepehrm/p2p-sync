package p2p

import "gitlab.com/thesepehrm/p2p-sync/common"

type Message uint8

const (
	StatusMsg Message = iota
	PingMsg
	GetKnownNodesMsg
	KnownNodesMsg

	NewDataMsg
	GetDataMsg
	DataMsg
)

type Packet interface {
	Name() string
	Type() Message
}

type PingPacket bool
type StatusPacket struct {
	NodeAddress   string
	KnownNodesNum int
}

type GetKnownNodesPacket int
type KnownNodesPacket map[common.Hash]string

type NewDataPacket common.Hash
type GetDataPacket common.Hash
type DataMsgPacket struct {
	Key   common.Hash
	Value string
}

func (*StatusPacket) Name() string  { return "Status" }
func (*StatusPacket) Type() Message { return StatusMsg }

func (*PingPacket) Name() string  { return "Ping" }
func (*PingPacket) Type() Message { return PingMsg }

func (*GetKnownNodesPacket) Name() string  { return "GetKnownNodes" }
func (*GetKnownNodesPacket) Type() Message { return GetKnownNodesMsg }

func (*KnownNodesPacket) Name() string  { return "KnownNodes" }
func (*KnownNodesPacket) Type() Message { return KnownNodesMsg }

func (*NewDataPacket) Name() string  { return "NewData" }
func (*NewDataPacket) Type() Message { return NewDataMsg }

func (*GetDataPacket) Name() string  { return "GetData" }
func (*GetDataPacket) Type() Message { return GetDataMsg }

func (*DataMsgPacket) Name() string  { return "Data" }
func (*DataMsgPacket) Type() Message { return DataMsg }
