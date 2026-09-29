package main

import (
	"fmt"

	"github.com/qimaotech/modu/internal/output"
	"github.com/spf13/cobra"
)

func runCheckout(cmd *cobra.Command, args []string) {
	eng := loadConfig()
	feature := args[0]
	display := startProgress("正在接手 " + feature)
	ctx := display.Context(cmd.Context())
	results, err := eng.CheckoutFeature(ctx, feature)
	display.Close()
	fmt.Print(output.New(outputFmt).FormatCheckoutResponse(feature, results, err))
	if err != nil {
		exitOperation(ctx)
	}
}
