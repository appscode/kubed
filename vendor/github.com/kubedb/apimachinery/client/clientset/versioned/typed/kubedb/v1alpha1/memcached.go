/*
Copyright 2018 The KubeDB Authors.

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

package v1alpha1

import (
	v1alpha1 "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	scheme "github.com/kubedb/apimachinery/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// MemcachedsGetter has a method to return a MemcachedInterface.
// A group's client should implement this interface.
type MemcachedsGetter interface {
	Memcacheds(namespace string) MemcachedInterface
}

// MemcachedInterface has methods to work with Memcached resources.
type MemcachedInterface interface {
	Create(*v1alpha1.Memcached) (*v1alpha1.Memcached, error)
	Update(*v1alpha1.Memcached) (*v1alpha1.Memcached, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Memcached, error)
	List(opts v1.ListOptions) (*v1alpha1.MemcachedList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Memcached, err error)
	MemcachedExpansion
}

// memcacheds implements MemcachedInterface
type memcacheds struct {
	client rest.Interface
	ns     string
}

// newMemcacheds returns a Memcacheds
func newMemcacheds(c *KubedbV1alpha1Client, namespace string) *memcacheds {
	return &memcacheds{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the memcached, and returns the corresponding memcached object, and an error if there is any.
func (c *memcacheds) Get(name string, options v1.GetOptions) (result *v1alpha1.Memcached, err error) {
	result = &v1alpha1.Memcached{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("memcacheds").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Memcacheds that match those selectors.
func (c *memcacheds) List(opts v1.ListOptions) (result *v1alpha1.MemcachedList, err error) {
	result = &v1alpha1.MemcachedList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("memcacheds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested memcacheds.
func (c *memcacheds) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("memcacheds").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a memcached and creates it.  Returns the server's representation of the memcached, and an error, if there is any.
func (c *memcacheds) Create(memcached *v1alpha1.Memcached) (result *v1alpha1.Memcached, err error) {
	result = &v1alpha1.Memcached{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("memcacheds").
		Body(memcached).
		Do().
		Into(result)
	return
}

// Update takes the representation of a memcached and updates it. Returns the server's representation of the memcached, and an error, if there is any.
func (c *memcacheds) Update(memcached *v1alpha1.Memcached) (result *v1alpha1.Memcached, err error) {
	result = &v1alpha1.Memcached{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("memcacheds").
		Name(memcached.Name).
		Body(memcached).
		Do().
		Into(result)
	return
}

// Delete takes name of the memcached and deletes it. Returns an error if one occurs.
func (c *memcacheds) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("memcacheds").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *memcacheds) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("memcacheds").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched memcached.
func (c *memcacheds) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Memcached, err error) {
	result = &v1alpha1.Memcached{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("memcacheds").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
