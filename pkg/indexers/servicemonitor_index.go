package indexers

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/appscode/go/arrays"
	"github.com/appscode/kubed/pkg/util"
	"github.com/appscode/log"
	"github.com/appscode/pat"
	"github.com/blevesearch/bleve"
	prom "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

type ServiceMonitorIndexer interface {
	Add(svcMonitor *prom.ServiceMonitor) error
	Delete(svcMonitor *prom.ServiceMonitor) error
	Update(old, new *prom.ServiceMonitor) error
	Key(meta metav1.ObjectMeta) []byte
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

var _ ServiceMonitorIndexer = &ServiceMonitorIndexerImpl{}

type ServiceMonitorIndexerImpl struct {
	kubeClient clientset.Interface
	index      bleve.Index
}

func (ri *ServiceMonitorIndexerImpl) Add(svcMonitor *prom.ServiceMonitor) error {
	log.Infof("New service: %v", svcMonitor.Name)
	log.V(5).Infof("Service details: %v", svcMonitor)

	svc, err := ri.serviceForServiceMonitors(svcMonitor)
	if err != nil {
		return err
	}

	for _, pod := range svc.Items {
		key := ri.Key(pod.ObjectMeta)
		raw, err := ri.index.GetInternal(key)
		if err != nil || len(raw) == 0 {
			data := prom.ServiceMonitorList{Items: []*prom.ServiceMonitor{svcMonitor}}
			raw, err := json.Marshal(data)
			if err != nil {
				return err
			}
			err = ri.index.SetInternal(key, raw)
			if err != nil {
				return err
			}
		} else {
			var data prom.ServiceMonitorList
			err := json.Unmarshal(raw, &data)
			if err != nil {
				return err
			}

			if found, _ := arrays.Contains(data.Items, svcMonitor); !found {
				data.Items = append(data.Items, svcMonitor)
				raw, err := json.Marshal(data)
				if err != nil {
					return err
				}
				err = ri.index.SetInternal(key, raw)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ri *ServiceMonitorIndexerImpl) Delete(svcMonitor *prom.ServiceMonitor) error {
	svc, err := ri.serviceForServiceMonitors(svcMonitor)
	if err != nil {
		return err
	}

	for _, pod := range svc.Items {
		key := ri.Key(pod.ObjectMeta)
		raw, err := ri.index.GetInternal(key)
		if err != nil {
			return err
		}
		if len(raw) > 0 {
			var data prom.ServiceMonitorList
			err := json.Unmarshal(raw, &data)
			if err != nil {
				return err
			}
			var monitors []*prom.ServiceMonitor
			for i, valueSvc := range data.Items {
				if ri.equal(svcMonitor, valueSvc) {
					monitors = append(data.Items[:i], data.Items[i+1:]...)
					break
				}
			}

			if len(monitors) == 0 {
				// Remove unnecessary index
				err = ri.index.DeleteInternal(key)
				if err != nil {
					return err
				}
			} else {
				raw, err := json.Marshal(prom.ServiceMonitorList{Items: monitors})
				if err != nil {
					return err
				}
				err = ri.index.SetInternal(key, raw)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ri *ServiceMonitorIndexerImpl) Update(old, new *prom.ServiceMonitor) error {
	if !reflect.DeepEqual(old.Spec.Selector, new.Spec.Selector) {
		// Only update if selector changes
		err := ri.Delete(old)
		if err != nil {
			return err
		}
		err = ri.Add(new)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ri *ServiceMonitorIndexerImpl) serviceForServiceMonitors(svcMonitor *prom.ServiceMonitor) (*apiv1.ServiceList, error) {
	selector, err := metav1.LabelSelectorAsSelector(&svcMonitor.Spec.Selector)
	if err != nil {
		return &apiv1.ServiceList{}, err
	}
	if svcMonitor.Spec.NamespaceSelector.Any {
		return ri.kubeClient.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})
	}

	list := &apiv1.ServiceList{Items: make([]apiv1.Service, 0)}
	for _, ns := range svcMonitor.Spec.NamespaceSelector.MatchNames {
		pods, err := ri.kubeClient.CoreV1().Services(ns).List(metav1.ListOptions{
			LabelSelector: selector.String(),
		})
		if err == nil {
			list.Items = append(list.Items, pods.Items...)
		}
	}
	return list, nil
}

func (ri *ServiceMonitorIndexerImpl) equal(a, b *prom.ServiceMonitor) bool {
	if a.Name == b.Name && a.Namespace == b.Namespace {
		return true
	}
	return false
}

func (ri *ServiceMonitorIndexerImpl) Key(meta metav1.ObjectMeta) []byte {
	return []byte(util.GetGroupVersionKind(&apiv1.Service{}).String() + "/" + meta.Namespace + "/" + meta.Name)
}

func (ri *ServiceMonitorIndexerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Infoln("Request received at", req.URL.Path)
	params, found := pat.FromContext(req.Context())
	if !found {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	namespace, name := params.Get(":namespace"), params.Get(":name")
	if len(namespace) > 0 && len(name) > 0 {
		key := ri.Key(v1.ObjectMeta{Name: name, Namespace: namespace})
		if val, err := ri.index.GetInternal(key); err == nil && len(val) > 0 {
			if err := json.NewEncoder(w).Encode(json.RawMessage(val)); err == nil {
				w.Header().Set("Content-Type", "application/json")
				return
			} else {
				http.Error(w, "Server error"+err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.NotFound(w, req)
		}
		return
	}
	http.Error(w, "Bad Request", http.StatusBadRequest)
}
