package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/trust"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// PullOptions defines what and how to pull
type PullOptions struct {
	remote string
	all    bool

	// 修改：添加-s，--simplify-image标记（flag）
	simp bool
	// 修改

	platform  string
	untrusted bool
}

// NewPullCommand creates a new `docker pull` command
func NewPullCommand(dockerCli command.Cli) *cobra.Command {
	var opts PullOptions

	cmd := &cobra.Command{
		Use:   "pull [OPTIONS] NAME[:TAG|@DIGEST]",
		Short: "Pull an image or a repository from a registry",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.remote = args[0]
			return RunPull(dockerCli, opts)
		},
	}

	flags := cmd.Flags()

	flags.BoolVarP(&opts.all, "all-tags", "a", false, "Download all tagged images in the repository")

	// 修改：添加-s，--simplify-image标记（flag）
	flags.BoolVarP(&opts.simp, "simplify-image", "s", false, "Simplify image")
	// 修改

	// 设置opts中platform元素，默认为""
	command.AddPlatformFlag(flags, &opts.platform)
	// 设置opts中untrusted元素，默认为false
	command.AddTrustVerificationFlags(flags, &opts.untrusted, dockerCli.ContentTrustEnabled())

	return cmd
}

// RunPull performs a pull against the engine based on the specified options
func RunPull(cli command.Cli, opts PullOptions) error {
	// 将remote字符串解析
	distributionRef, err := reference.ParseNormalizedNamed(opts.remote)

	switch {
	case err != nil:
		return err
	// 判断 -a 标记和参数的关系
	// -a 参数不能和tag参数同时出现
	// 没有-a参数时也没有tag参数时，补全tag参数
	case opts.all && !reference.IsNameOnly(distributionRef):
		return errors.New("tag can't be used with --all-tags/-a")
	case !opts.all && reference.IsNameOnly(distributionRef):
		distributionRef = reference.TagNameOnly(distributionRef)
		if tagged, ok := distributionRef.(reference.Tagged); ok {
			fmt.Fprintf(cli.Out(), "Using default tag: %s\n", tagged.Tag())
		}
	}

	// 空context变量
	ctx := context.Background()

	// 认证镜像信息
	imgRefAndAuth, err := trust.GetImageReferencesAndAuth(ctx, nil, AuthResolver(cli), distributionRef.String())
	if err != nil {
		return err
	}

	// Check if reference has a digest
	_, isCanonical := distributionRef.(reference.Canonical)

	// 镜像校验默认选择，所以一般执行else中指令
	if !opts.untrusted && !isCanonical {
		err = trustedPull(ctx, cli, imgRefAndAuth, opts.platform)
	} else {
		// 修改：添加传递opts.simp参数
		err = imagePullPrivileged(ctx, cli, imgRefAndAuth, opts.all, opts.simp, opts.platform)
		// 修改
	}

	if err != nil {
		if strings.Contains(err.Error(), "when fetching 'plugin'") {
			return errors.New(err.Error() + " - Use `docker plugin install`")
		}
		return err
	}

	return nil
}
