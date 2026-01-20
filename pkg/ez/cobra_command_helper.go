package ez

import (
	"github.com/spf13/cobra"
)

type ICobraCommandHelper interface {
	GetBoolFlag(name string) bool
	GetStringFlag(name string) string
	GetIntFlag(name string) int
	IsVerbose() bool
	IsDryRun() bool
}

type CobraCommandHelperImpl struct {
	Command *cobra.Command
}

func (e *CobraCommandHelperImpl) GetBoolFlag(name string) bool {

	val, err := e.Command.Flags().GetBool(name)

	if err != nil {
		panic(err)
	}

	return val
}

func (e *CobraCommandHelperImpl) GetIntFlag(name string) int {

	val, err := e.Command.Flags().GetInt(name)

	if err != nil {
		panic(err)
	}

	return val
}

func (e *CobraCommandHelperImpl) IsVerbose() bool {
	val, err := e.Command.Flags().GetBool("verbose")
	if err != nil {
		panic(err)
	}
	return val
}

func (e *CobraCommandHelperImpl) IsDryRun() bool {
	val, err := e.Command.Flags().GetBool("dry-run")
	if err != nil {
		panic(err)
	}
	return val
}

func (e *CobraCommandHelperImpl) GetStringFlag(name string) string {

	val, err := e.Command.Flags().GetString(name)

	if err != nil {
		panic(err)
	}

	return val
}

type ContextKey string

const AppCtxKey = ContextKey("appContext")

func CommandHelper(cmd *cobra.Command) *CobraCommandHelperImpl {
	val := cmd.Context().Value(AppCtxKey)
	if appCtx, ok := val.(*CobraCommandHelperImpl); ok {
		appCtx.Command = cmd
		return appCtx
	}
	return nil
}
