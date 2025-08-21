package jsutils

import (
	"github.com/google/uuid"
	"github.com/torloj/easykube/ekctx"
)

type IUtils interface {
	UUID() string
}

type Utils struct {
	EKContext *ekctx.EKContext
}

func NewUtils(ctx *ekctx.EKContext) IUtils {
	return &Utils{EKContext: ctx}
}

func (u *Utils) UUID() string {
	return uuid.New().String()
}
