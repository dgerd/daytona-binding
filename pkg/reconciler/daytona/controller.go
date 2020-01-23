/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package daytona

import (
	"context"

	dbinformer "github.com/dgerd/daytona-binding/pkg/client/injection/informers/daytonabinding/v1alpha1/daytonabinding"
	"knative.dev/pkg/client/injection/ducks/duck/v1/podable"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection/clients/dynamicclient"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
	"knative.dev/pkg/webhook/podbinding"

	"github.com/dgerd/daytona-binding/pkg/apis/daytonabinding/v1alpha1"
)

const (
	controllerAgentName = "daytona-controller"
)

// NewController returns a new DaytonaBinding reconciler. This reconciler tracks changes on the
// DaytonaBinding CRD object to ensure the webhook is targeting the current tracked resources and
// that the latest configuration options are being applied.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	dbInformer := dbinformer.Get(ctx)
	dc := dynamicclient.Get(ctx)
	podInformerFactory := podable.Get(ctx)

	c := &podbinding.BaseReconciler{
		GVR: v1alpha1.SchemeGroupVersion.WithResource("daytonabindings"),
		Get: func(namespace string, name string) (podbinding.Bindable, error) {
			return dbInformer.Lister().DaytonaBindings(namespace).Get(name)
		},
		DynamicClient: dc,
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
	}
	impl := controller.NewImpl(c, logger, "DaytonaBindings")

	logger.Info("Setting up event handlers")

	dbInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	c.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))
	c.Factory = &duck.CachedInformerFactory{
		Delegate: &duck.EnqueueInformerFactory{
			Delegate:     podInformerFactory,
			EventHandler: controller.HandleAll(c.Tracker.OnChanged),
		},
	}

	return impl
}

func ListAll(ctx context.Context, handler cache.ResourceEventHandler) podbinding.ListAll {
	dbInformer := dbinformer.Get(ctx)

	// Whenever a DaytonaBinding changes our webhook programming might change.
	dbInformer.Informer().AddEventHandler(handler)

	return func() ([]podbinding.Bindable, error) {
		l, err := dbInformer.Lister().List(labels.Everything())
		if err != nil {
			return nil, err
		}
		bl := make([]podbinding.Bindable, 0, len(l))
		for _, elt := range l {
			bl = append(bl, elt)
		}
		return bl, nil
	}
}
