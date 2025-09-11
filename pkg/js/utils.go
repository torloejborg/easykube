package jsutils

import (
	"github.com/google/uuid"
	"github.com/torloejborg/easykube/cmd"
)

type IUtils interface {
	UUID() string
}

type Utils struct {
	EKContext *cmd.CobraCommandHelperImpl
}

func NewUtils(ctx *cmd.CobraCommandHelperImpl) IUtils {
	return &Utils{EKContext: ctx}
}

func (u *Utils) UUID() string {
	return uuid.New().String()
}
