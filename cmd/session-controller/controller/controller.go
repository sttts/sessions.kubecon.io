package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	clientset "k8s.io/kubecon/pkg/client/clientset/versioned"
	informers "k8s.io/kubecon/pkg/client/informers/externalversions"
	listers "k8s.io/kubecon/pkg/client/listers/kubecon.io/v1"
)

// Controller is the controller implementation for Session resources
type Controller struct {
	kubeconClient clientset.Interface

	sessionLister  listers.SessionLister
	sessionsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

// NewController returns a new session controller
func NewController(
	kubeconClientset clientset.Interface,
	kubeconInformerFactory informers.SharedInformerFactory) *Controller {
	sessionInformer := kubeconInformerFactory.Kubecon().V1().Sessions()

	controller := &Controller{
		kubeconClient:  kubeconClientset,
		sessionLister:  sessionInformer.Lister(),
		sessionsSynced: sessionInformer.Informer().HasSynced,
		workqueue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Sessions"),
	}

	sessionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSession,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueSession(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Session controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.sessionsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Session resource with this namespace/name
	session, err := c.sessionLister.Sessions(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("session '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if session.Spec.Capacity == session.Status.Attendees {
		// nothing to do, dequeue
		return nil
	}

	if session.Spec.Capacity < session.Status.Attendees {
		session.Status.Attendees--
	} else {
		session.Status.Attendees++
	}

	if _, err := c.kubeconClient.KubeconV1().Sessions(session.Namespace).Update(session); errors.IsConflict(err) {
		c.workqueue.Add(key)
		return err
	} else if err != nil {
		return err
	}

	// no need to requeue, we will see our own update.

	return nil
}

// enqueueSession takes a Session resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueSession(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddAfter(key, time.Second)
}
