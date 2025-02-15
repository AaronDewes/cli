package typescript

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/supabase/cli/internal/utils"
)

var ctx = context.Background()

func Run(useLocal bool, dbUrl string) error {
	if useLocal && dbUrl != "" {
		return errors.New("Cannot specify both --local and --db-url")
	} else if !useLocal && dbUrl == "" {
		return errors.New("Must specify either --local or --db-url")
	}

	if err := utils.LoadConfig(); err != nil {
		return err
	}

	if useLocal {
		if err := utils.AssertSupabaseStartIsRunning(); err != nil {
			return err
		}

		exec, err := utils.Docker.ContainerExecCreate(
			ctx,
			utils.PgmetaId,
			types.ExecConfig{
				Env: []string{
					"PG_META_DB_HOST=" + utils.DbId,
				},
				Cmd: []string{
					"node", "bin/src/server/app.js", "gen", "types", "typescript", "--include-schemas", strings.Join(append([]string{"public"}, utils.Config.Api.Schemas...), ","),
				},
				AttachStderr: true,
				AttachStdout: true,
			},
		)
		if err != nil {
			return err
		}

		resp, err := utils.Docker.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
		if err != nil {
			return err
		}
		var genBuf, errBuf bytes.Buffer
		if _, err := stdcopy.StdCopy(&genBuf, &errBuf, resp.Reader); err != nil {
			return err
		}
		if errBuf.Len() > 0 {
			return errors.New(errBuf.String())
		}

		fmt.Print(genBuf.String())
		return nil
	}

	// run typegen on the dbUrl
	{
		if err := utils.AssertDockerIsRunning(); err != nil {
			return err
		}

		defer utils.DockerRemoveAll()

		metaImage := utils.GetRegistryImageUrl(utils.PgmetaImage)
		if _, _, err := utils.Docker.ImageInspectWithRaw(ctx, metaImage); err != nil {
			fmt.Fprintln(os.Stderr, "Downloading type generator...")
			out, err := utils.Docker.ImagePull(ctx, metaImage, types.ImagePullOptions{})
			if err != nil {
				return err
			}
			if _, err := io.ReadAll(out); err != nil {
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "Done downloading type generator")
		}

		out, err := utils.DockerRun(
			ctx,
			"supabase_gen_types_typescript",
			&container.Config{
				Image: utils.PgmetaImage,
				Env: []string{
					"PG_META_DB_URL=" + dbUrl,
				},
				Cmd: []string{
					"node", "bin/src/server/app.js", "gen", "types", "typescript", "--include-schemas", strings.Join(append([]string{"public"}, utils.Config.Api.Schemas...), ","),
				},
			},
			&container.HostConfig{},
		)
		if err != nil {
			return err
		}
		var genBuf, errBuf bytes.Buffer
		if _, err := stdcopy.StdCopy(&genBuf, &errBuf, out); err != nil {
			return err
		}
		if errBuf.Len() > 0 {
			return errors.New(errBuf.String())
		}

		fmt.Print(genBuf.String())
	}

	return nil
}
