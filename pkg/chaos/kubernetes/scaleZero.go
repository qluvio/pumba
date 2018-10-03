package kubernetes

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/alexei-led/pumba/pkg/chaos"
	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/alexei-led/pumba/pkg/util"
)

// ScaleToZeroCommand command
type ScaleToZeroCommand struct {
	kubeIface dynamic.Interface
	kind      string
	namespace string
	names     []string
	pattern   string
	duration  time.Duration
	dryRun    bool
}

// NewScaleToZeroCommand create new Patch command instance
func NewScaleToZeroCommand(kubeIface interface{}, kind, namespace string, names []string, pattern string, intervalStr, durationStr string, dryRun bool) (chaos.Command, error) {
	// get interval
	interval, err := util.GetIntervalValue(intervalStr)
	if err != nil {
		return nil, err
	}
	// get duration
	duration, err := util.GetDurationValue(durationStr, interval)
	if err != nil {
		return nil, err
	}
	command := &ScaleToZeroCommand{kubeIface.(dynamic.Interface), kind, namespace, names, pattern, duration, dryRun}
	return command, nil
}

//  patchStringValue specifies a patch operation for a uint32.
type patchUInt32Value struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value uint32 `json:"value"`
}

// Run scale to zero command
func (c *ScaleToZeroCommand) Run(ctx context.Context, random bool) error {
	log.Debug("setting replicas to zero for matching resources")
	log.WithFields(log.Fields{
		"namespace": c.namespace,
	}).Debug("listing matching kube resources")

	var deploymentsResource = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	resourceIface := c.kubeIface.Resource(deploymentsResource)
	list, err := resourceIface.List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("failed to list resources")
		return err
	}

	resources := list.Items
	if len(resources) == 0 {
		log.Warning("no resource to scale to zero")
		return nil
	}

	// select single random container from matching container and replace list with selected item
	if random {
		log.Debug("selecting single random resource")
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		resources = []unstructured.Unstructured{list.Items[r.Intn(len(resources))]}
	}

	// update scale for selected resources
	var cancels []context.CancelFunc
	for _, resource := range resources {
		name := resource.GetName()
		log.WithFields(log.Fields{
			"name":      name,
			"namespace": resource.GetNamespace(),
		}).Debug("setting number of replicas to zero for resource")

		patchCtx, cancel := context.WithCancel(ctx)
		cancels = append(cancels, cancel)

		payload := []patchUInt32Value{{
			Op:    "replace",
			Path:  "/spec/replicas",
			Value: 0,
		}}
		patchBytes, _ := json.Marshal(payload)
		payload = []patchUInt32Value{{
			Op:    "replace",
			Path:  "/spec/replicas",
			Value: 1,
		}}
		revertBytes, _ := json.Marshal(payload)

		err = patchResource(patchCtx, resourceIface, name, patchBytes, revertBytes, c.duration, c.dryRun)
		if err != nil {
			log.WithError(err).Error("failed to set replicas to zero")
			// break on error - to cancel all open contexts and avoid go routine leaks
			break
		}
	}

	// cancel context to avoid leaks
	defer func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()

	return err
}
