package client // import "github.com/docker/docker/client"

import (
	"context"
	"net/url"

	"github.com/docker/docker/api/types"
)

// ContainerStart sends a request to the docker daemon to start a container.
func (cli *Client) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	query := url.Values{}
	if len(options.CheckpointID) != 0 {
		query.Set("checkpoint", options.CheckpointID)
	}
	if len(options.CheckpointDir) != 0 {
		query.Set("checkpoint-dir", options.CheckpointDir)
	}

	// 修改： 添加simplify-image参数到查询中
	if options.Simp {
		query.Set("simplify-image", "yes")
	}
	// 修改

	resp, err := cli.post(ctx, "/containers/"+containerID+"/start", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
