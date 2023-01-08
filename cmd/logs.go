package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cli, err := client.NewEnvClient()

		if err != nil {
			panic(err)
		}

		// todo --tail flag追加する
		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			// All: true,
		})
		containers_map := map[string]string{}
		var containers_name_slice []string
		for _, v := range containers {
			for _, a := range v.Names {
				containers_map[strings.TrimLeft(a, "/")+"("+v.Image+")"] = v.ID
				containers_name_slice = append(containers_name_slice, strings.TrimLeft(a, "/")+"("+v.Image+")")
			}
		}

		selected_container := chooseValueFromPromptItems("Select Container", containers_name_slice)

		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       "all",
			Details:    true,
		}

		out, err := cli.ContainerLogs(ctx, containers_map[selected_container], options)
		if err != nil {
			panic(err)
		}
		// https://matsuand.github.io/docs.docker.jp.onthefly/engine/api/sdk/examples/
		// io.Copy(os.Stdout, out)
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
