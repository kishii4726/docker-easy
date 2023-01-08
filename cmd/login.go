package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func chooseValueFromPromptItems(s string, l []string) string {
	prompt := promptui.Select{
		Label: s,
		Items: l,
	}
	_, v, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	return v
}

var loginCmd = &cobra.Command{
	Use:   "login",
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

		containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
			// All: true,
		})
		if err != nil {
			panic(err)
		}
		containers_map := map[string]string{}
		var containers_name_slice []string
		for _, container := range containers {
			for _, a := range container.Names {
				containers_map[strings.TrimLeft(a, "/")+"("+container.Image+")"] = container.ID
				containers_name_slice = append(containers_name_slice, strings.TrimLeft(a, "/")+"("+container.Image+")")
			}
		}

		selected_container := chooseValueFromPromptItems("Select Container", containers_name_slice)

		selected_shell := chooseValueFromPromptItems("Select Shell", []string{"/bin/sh", "/bin/bash"})

		// fmt.Println(selected_container + ": " + containers_map[selected_container])
		// fmt.Println(selected_shell)
		execOpts := types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Cmd:          []string{selected_shell},
		}

		resp, err := cli.ContainerExecCreate(context.Background(), containers_map[selected_container], execOpts)
		if err != nil {
			panic(err)
		}

		respTwo, err := cli.ContainerExecAttach(context.Background(), resp.ID, types.ExecStartCheck{})
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := respTwo.Conn.Close(); err != nil {
				log.Panic(err)
			}
			log.Println("connection closed")
		}()

		fd := int(os.Stdin.Fd())
		if terminal.IsTerminal(fd) {
			state, err := terminal.MakeRaw(fd)
			if err != nil {
				log.Panic(err)
			}
			defer terminal.Restore(fd, state)

			// w, h, err := terminal.GetSize(fd)
			// if err != nil {
			// 	log.Panic(err)
			// }
			// if w < 0 || h < 0 {
			// 	log.Panic("terminal size error", " w: ", w, " h:", h)
			// }
			// resizeOptions := types.ResizeOptions{
			// 	Height: uint(h),
			// 	Width:  uint(w),
			// }
			// if err := cli.ContainerExecResize(ctx, resp.ID, resizeOptions); err != nil {
			// 	log.Panic(err)
			// }
		}

		go io.Copy(respTwo.Conn, os.Stdin)
		stdcopy.StdCopy(os.Stdout, os.Stderr, respTwo.Reader)

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

}

// https://haibara-works.hatenablog.com/entry/2020/12/05/235227
// https://matsuand.github.io/docs.docker.jp.onthefly/engine/api/sdk/examples/
