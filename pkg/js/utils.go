package jsutils

import (
	"github.com/google/uuid"
	"github.com/torloejborg/easykube/pkg/ez"
)

type IUtils interface {
	UUID() string
}

type Utils struct {
	CommandHelper *ez.CobraCommandHelperImpl
}

func NewUtils(commandHelper *ez.CobraCommandHelperImpl) IUtils {
	return &Utils{CommandHelper: commandHelper}
}

func (u *Utils) UUID() string {
	return uuid.New().String()
}
