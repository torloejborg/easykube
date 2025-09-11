package jsutils

import (
	"github.com/google/uuid"
	"github.com/torloejborg/easykube/pkg/ez"
)

type IUtils interface {
	UUID() string
}

type Utils struct {
	EKContext *ez.CobraCommandHelperImpl
}

func NewUtils(ctx *ez.CobraCommandHelperImpl) IUtils {
	return &Utils{EKContext: ctx}
}

func (u *Utils) UUID() string {
	return uuid.New().String()
}
