package ekctx

import (
	"log"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type EKContext struct {
	Logger  *log.Logger
	Printer *Printer
	Command *cobra.Command
	Fs      afero.Fs
}

func (e *EKContext) GetBoolFlag(name string) bool {

	val, err := e.Command.Flags().GetBool(name)

	if err != nil {
		panic(err)
	}

	return val
}

func (e *EKContext) GetStringFlag(name string) string {

	val, err := e.Command.Flags().GetString(name)

	if err != nil {
		panic(err)
	}

	return val
}

type ContextKey string

const AppCtxKey = ContextKey("appContext")

func GetAppContext(cmd *cobra.Command) *EKContext {
	val := cmd.Context().Value(AppCtxKey)
	if appCtx, ok := val.(*EKContext); ok {
		appCtx.Command = cmd
		return appCtx
	}
	return nil
}
