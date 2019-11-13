package fetcp

type Packet interface {
	Serialize() []byte
}

type Protocol interface {
	ReadPacket(c *Conn) (Packet, error)
}
