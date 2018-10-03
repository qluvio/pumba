package kubernetes

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// update k8s resource state, restore state on timeout or abort
func patchResource(ctx context.Context, resourceIface dynamic.NamespaceableResourceInterface, name string, patchBytes []byte, revertBytes []byte, duration time.Duration, dryRun bool) error {
	log.WithFields(log.Fields{
		"name": name,
	}).Debug("patching resources")

	// patch resource
	_, err := resourceIface.Patch(name, types.JSONPatchType, patchBytes)
	if err != nil {
		log.WithError(err).Error("failed to patch the resource")
		return err
	}

	// create new context with timeout for canceling
	stopCtx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	// wait for specified duration and then try to restore previous state (where it applied) or stop on ctx.Done()
	select {
	case <-ctx.Done():
		log.WithFields(log.Fields{
			"name": name,
		}).Debug("restoring previous state on abort")
		// use different context to revert patch netem since parent context is canceled
		_, err = resourceIface.Patch(name, types.JSONPatchType, revertBytes)
	case <-stopCtx.Done():
		log.WithFields(log.Fields{
			"name": name,
		}).Debug("resstore previous state on timout")
		// use parent context to revert patch in container
		_, err = resourceIface.Patch(name, types.JSONPatchType, revertBytes)
	}
	return err
}
