package operator

import (
	"errors"
	"reflect"

	acrt "github.com/appscode/go/runtime"
	"github.com/appscode/kubed/pkg/util"
	"github.com/appscode/log"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

// Blocks caller. Intended to be called as a Go routine.
func (op *Operator) WatchServiceMonitors() {
	if !util.IsPreferredAPIResource(op.KubeClient, prom.TPRGroup+"/"+prom.TPRVersion, prom.TPRServiceMonitorsKind) {
		log.Warningf("Skipping watching non-preferred GroupVersion:%s Kind:%s", prom.TPRGroup+"/"+prom.TPRVersion, prom.TPRServiceMonitorsKind)
		return
	}

	defer acrt.HandleCrash()

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return op.PromClient.ServiceMonitors(apiv1.NamespaceAll).List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return op.PromClient.ServiceMonitors(apiv1.NamespaceAll).Watch(metav1.ListOptions{})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&prom.ServiceMonitor{},
		op.syncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				if res, ok := obj.(*prom.ServiceMonitor); ok {
					log.Infof("ServiceMonitor %s@%s added", res.Name, res.Namespace)
					util.AssignTypeKind(res)

					if op.Opt.EnableSearchIndex {
						if err := op.SearchIndex.HandleAdd(obj); err != nil {
							log.Errorln(err)
						}
					}

					if op.Opt.EnableReverseIndex {
						if err := op.ReverseIndex.ServiceMonitor.Add(res); err != nil {
							log.Errorln(err)
						}
						if op.ReverseIndex.Prometheus != nil {
							proms, err := op.PromClient.Prometheuses(apiv1.NamespaceAll).List(metav1.ListOptions{})
							if err != nil {
								log.Errorln(err)
								return
							}
							if promList, ok := proms.(*prom.PrometheusList); ok {
								op.ReverseIndex.Prometheus.AddServiceMonitor(res, promList.Items)
							}
						}
					}
				}
			},
			DeleteFunc: func(obj interface{}) {
				if res, ok := obj.(*prom.ServiceMonitor); ok {
					log.Infof("ServiceMonitor %s@%s deleted", res.Name, res.Namespace)
					util.AssignTypeKind(res)

					if op.Opt.EnableSearchIndex {
						if err := op.SearchIndex.HandleDelete(obj); err != nil {
							log.Errorln(err)
						}
					}
					if op.TrashCan != nil {
						op.TrashCan.Delete(res.TypeMeta, res.ObjectMeta, obj)
					}

					if op.Opt.EnableReverseIndex {
						if err := op.ReverseIndex.ServiceMonitor.Delete(res); err != nil {
							log.Errorln(err)
						}
						if op.ReverseIndex.Prometheus != nil {
							op.ReverseIndex.Prometheus.DeleteServiceMonitor(res)
						}
					}
				}
			},
			UpdateFunc: func(old, new interface{}) {
				oldRes, ok := old.(*prom.ServiceMonitor)
				if !ok {
					log.Errorln(errors.New("Invalid ServiceMonitor object"))
					return
				}
				newRes, ok := new.(*prom.ServiceMonitor)
				if !ok {
					log.Errorln(errors.New("Invalid ServiceMonitor object"))
					return
				}
				util.AssignTypeKind(oldRes)
				util.AssignTypeKind(newRes)

				if op.Opt.EnableSearchIndex {
					op.SearchIndex.HandleUpdate(old, new)
				}
				if op.TrashCan != nil && op.Config.RecycleBin.HandleUpdates {
					if !reflect.DeepEqual(oldRes.Labels, newRes.Labels) ||
						!reflect.DeepEqual(oldRes.Annotations, newRes.Annotations) ||
						!reflect.DeepEqual(oldRes.Spec, newRes.Spec) {
						op.TrashCan.Update(newRes.TypeMeta, newRes.ObjectMeta, old, new)
					}
				}

				if op.Opt.EnableReverseIndex {
					if err := op.ReverseIndex.ServiceMonitor.Update(oldRes, newRes); err != nil {
						log.Errorln(err)
					}
				}
			},
		},
	)
	ctrl.Run(wait.NeverStop)
}
