package cluster

import (
	"context"

	"github.com/tinylib/msgp/msgp"
)

type Node interface {
	IsLocal() bool
	IsReady() bool
	GetPartitions() []int32
	GetPriority() int
	Post(context.Context, string, string, Traceable) ([]byte, error)
	PostInto(result msgp.Decodable, ctx context.Context, name, path string, body Traceable) error
	GetName() string
}
