package cmd

import (
	"log"

	"git.fogcdn.top/tcgroup/tool/service/iam"
	"github.com/spf13/cobra"
)

var cmdIam = &cobra.Command{
	Use:   "iam",
	Short: "gen iam sql",
	Long:  "gen iam sql\n命令参数\t\t例子\n1.xlsx文件路径\t\thost_group_rename.xlsx\n2.sheet名\t\thost_group_sheet\n3.原主机组编码列\t5\n4.新主机组编码列\t6",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		inFilePath := args[0]
		outFilePath := "iam.sql"
		log.Printf("gen iam sql start, inFilePath:%s", inFilePath)
		iam.GenIamSql(inFilePath, outFilePath)
	},
}

func init() {
	RootCmd.AddCommand(cmdIam)
}
