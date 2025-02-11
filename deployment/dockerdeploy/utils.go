package dockerdeploy

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func dockerBuildAndPipe(ctx context.Context, logger *zap.Logger, cli *client.Client, buildContext io.Reader, options types.ImageBuildOptions) error {
	buildResp, err := cli.ImageBuild(ctx, buildContext, options)
	if err != nil {
		return errors.Wrap(err, "failed to build image")
	}
	defer buildResp.Body.Close()

	pipeRdr, pipeWrt := io.Pipe()
	defer pipeWrt.Close()
	go func() {
		scanner := bufio.NewScanner(pipeRdr)
		for scanner.Scan() {
			logger.Debug("docker build output", zap.String("text", scanner.Text()))
		}
	}()

	dec := json.NewDecoder(buildResp.Body)
	for dec.More() {
		var streamMsg struct {
			Stream string `json:"stream"`
			Error  string `json:"error"`
		}
		err := dec.Decode(&streamMsg)
		if err != nil {
			return errors.Wrap(err, "json decode failure while reading build progress")
		}

		if streamMsg.Error != "" {
			return errors.Wrap(errors.New(streamMsg.Error), "failed during image building")
		}

		pipeWrt.Write([]byte(streamMsg.Stream))
	}

	return nil
}

func dockerPullAndPipe(ctx context.Context, logger *zap.Logger, cli *client.Client, refStr string, options types.ImagePullOptions) error {
	pullResp, err := cli.ImagePull(ctx, refStr, options)
	if err != nil {
		return errors.Wrap(err, "failed to pull image")
	}
	defer pullResp.Close()

	dec := json.NewDecoder(pullResp)

	for dec.More() {
		var streamMsg struct {
			Status string `json:"status"`
		}
		err := dec.Decode(&streamMsg)
		if err != nil {
			return errors.Wrap(err, "json decode failure while reading pull progress")
		}

		switch streamMsg.Status {
		case "Waiting":
		case "Downloading":
		case "Extracting":
		default:
			logger.Debug("docker pull output", zap.String("text", streamMsg.Status))
		}
	}

	return nil
}
