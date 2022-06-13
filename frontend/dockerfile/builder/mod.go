package builder

import (
	"context"

	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/util/bklog"
)

func loadMods(ctx context.Context, c client.Client, modName string, opts map[string]string, ref client.Reference) ([]byte, error) {
	modArg := opts[keyModArg]
	if modArg == "" {
		modArg = "dockerfile-mod"
	}
	modSt, _, _, err := contextByName(ctx, c, modArg, nil, nil)
	if err != nil {
		return nil, err
	}

	var dtDockerfileMod []byte

	if modSt != nil {
		modDef, err := modSt.Marshal(ctx)
		if err != nil {
			return nil, err
		}
		modRes, err := c.Solve(ctx, client.SolveRequest{
			Definition: modDef.ToPB(),
		})
		if err != nil {
			return nil, err
		}

		modRef, err := modRes.SingleRef()
		if err != nil {
			return nil, err
		}
		dtDockerfileMod, err = modRef.ReadFile(ctx, client.ReadRequest{
			Filename: modName,
		})
		bklog.G(ctx).Infof("mod: %s, data: %s", modName, string(dtDockerfileMod))
		if err != nil {
			return nil, err
		}
	}

	if dtDockerfileMod == nil {
		// Look for a mod in the main context
		if _, err := ref.StatFile(ctx, client.StatRequest{Path: modName}); err == nil {
			dtDockerfileMod, err = ref.ReadFile(ctx, client.ReadRequest{
				Filename: modName,
			})
			bklog.G(ctx).Infof("mod: %s, data: %s", modName, string(dtDockerfileMod))
			if err != nil {
				return nil, err
			}
		}
	}

	return dtDockerfileMod, nil
}
