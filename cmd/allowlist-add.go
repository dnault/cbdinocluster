package cmd

import (
	"github.com/couchbaselabs/cbdinocluster/deployment/clouddeploy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var allowListAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"create"},
	Short:   "Adds an allowed CIDRs",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		helper := CmdHelper{}
		logger := helper.GetLogger()
		ctx := helper.GetContext()

		deployer, cluster, err := helper.IdentifyCluster(ctx, args[0])
		if err != nil {
			logger.Fatal("failed to identify cluster", zap.Error(err))
		}

		cloudDeployer, ok := deployer.(*clouddeploy.Deployer)
		if !ok {
			logger.Fatal("allow-lists are only supported for cloud deployer")
		}

		err = cloudDeployer.AddAllowListEntry(ctx, cluster.GetID(), args[1])
		if err != nil {
			logger.Fatal("failed to add allow list entry", zap.Error(err))
		}
	},
}

func init() {
	allowListCmd.AddCommand(allowListAddCmd)
}
