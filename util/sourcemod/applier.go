package sourcemod

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/util/bklog"
	binfotypes "github.com/moby/buildkit/util/buildinfo/types"
)

type Mod struct {
	Sources []Source
}

// 1. Check for duplicates
// 2. Check that the ref is fully qualified
// 3. Check that the replacement is fully qualified
func (m Mod) validate() error {
	if len(m.Sources) == 0 {
		return errors.New("no sources")
	}

	checked := make(map[string]struct{})
	for _, s := range m.Sources {
		if _, exists := checked[s.Ref]; exists {
			return fmt.Errorf("duplicate source ref: %s", s.Ref)
		}
		if err := validateRef(s.Ref); err != nil {
			return err
		}

		if s.Replace != "" {
			if err := validateRef(s.Replace); err != nil {
				return err
			}
		}

		checked[s.Ref] = struct{}{}
	}

	return nil
}

func validateRef(s string) error {
	ref, err := reference.ParseNormalizedNamed(s)
	if err != nil {
		return err
	}

	if reference.IsNameOnly(ref) {
		return fmt.Errorf("missing tag for reference: %s", s)
	}

	return nil
}

type Source struct {
	Type    binfotypes.SourceType
	Ref     string
	Replace string
}

type Applier struct {
	mod Mod

	validatedMu   sync.Mutex
	validated     bool
	validationErr error
}

func NewApplier(mod Mod) *Applier {
	return &Applier{
		mod: mod,
	}
}

func (a *Applier) Validate() error {
	a.validatedMu.Lock()
	defer a.validatedMu.Unlock()

	if a.validated {
		return a.validationErr
	}

	a.validationErr = a.mod.validate()
	a.validated = true
	return a.validationErr
}

func (a *Applier) UpdateRef(ctx context.Context, ref string) (string, error) {
	logger := bklog.G(ctx).WithField("ref", ref)
	if err := a.Validate(); err != nil {
		return "", err
	}

	parsed, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return "", err
	}

	// Make sure there is a tag on this
	parsed = reference.TagNameOnly(parsed)

	for _, source := range a.mod.Sources {
		if source.Ref != parsed.String() {
			logger.Infof("No match: %s!=%s", parsed, source.Ref)
			continue
		}

		if source.Replace != "" {
			logger.Infof("Resolving as replacement: %s", source.Replace)
			return source.Replace, nil
		}
		logger.Infof("Resolving as normal: %s", source.Ref)
		return source.Ref, nil
	}
	return "", fmt.Errorf("ref is not present in sources list: %s", ref)
}

// func (a *Applier) ResolveImageConfig(ctx context.Context, ref string, opt llb.ResolveImageConfigOpt) (digest.Digest, []byte, error) {
// 	logger := bklog.G(ctx).WithField("ref", ref)
// 	logger.Info("Resolving")
// 	if err := a.validate(); err != nil {
// 		return "", nil, err
// 	}
//
// 	parsed, err := reference.ParseNormalizedNamed(ref)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	for _, source := range a.mod.Sources {
// 		if source.Type != binfotypes.SourceTypeDockerImage {
// 			logger.Info("Not a docker image")
// 			continue
// 		}
//
// 		if source.Ref != parsed.String() {
// 			logger.Infof("No match: %s!=%s", parsed, source.Ref)
// 			continue
// 		}
//
// 		if source.Replace != "" {
// 			logger.Infof("Resolving as replacement: %s", source.Replace)
// 			return a.resolver.ResolveImageConfig(ctx, source.Replace, opt)
// 		}
// 		logger.Infof("Resolving as normal: %s", source.Ref)
// 		return a.ResolveImageConfig(ctx, source.Ref, opt)
// 	}
//
// 	return "", nil, fmt.Errorf("ref is not present in sources list: %s", ref)
// }
