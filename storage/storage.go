package storage

type Kind int

const (
	MemKind   Kind = iota // mem
	MongoKind             // mongo
)
