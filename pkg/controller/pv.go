package controller

import (
	acrt "github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/util"
	"github.com/appscode/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

// Blocks caller. Intended to be called as a Go routine.
func (c *Controller) WatchPersistentVolumes() {
	if !util.IsPreferredAPIResource(c.KubeClient, apiv1.SchemeGroupVersion.String(), "PersistentVolume") {
		log.Warningf("Skipping watching non-preferred GroupVersion:%s Kind:%s", apiv1.SchemeGroupVersion.String(), "PersistentVolume")
		return
	}

	defer acrt.HandleCrash()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return c.KubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.KubeClient.CoreV1().PersistentVolumes().Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&apiv1.PersistentVolume{},
		c.SyncPeriod,
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if pv, ok := obj.(*apiv1.PersistentVolume); ok {
					log.Infof("PersistentVolume %s@%s deleted", pv.Name, pv.Namespace)
					c.Saver.Save(pv.ObjectMeta, obj)
				}
			},
		},
	)
	ctrl.Run(wait.NeverStop)
}
