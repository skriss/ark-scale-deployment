/*
Copyright 2018 the Heptio Ark contributors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/heptio/ark/pkg/apis/ark/v1"
	scheme "github.com/heptio/ark/pkg/generated/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PodVolumeRestoresGetter has a method to return a PodVolumeRestoreInterface.
// A group's client should implement this interface.
type PodVolumeRestoresGetter interface {
	PodVolumeRestores(namespace string) PodVolumeRestoreInterface
}

// PodVolumeRestoreInterface has methods to work with PodVolumeRestore resources.
type PodVolumeRestoreInterface interface {
	Create(*v1.PodVolumeRestore) (*v1.PodVolumeRestore, error)
	Update(*v1.PodVolumeRestore) (*v1.PodVolumeRestore, error)
	UpdateStatus(*v1.PodVolumeRestore) (*v1.PodVolumeRestore, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.PodVolumeRestore, error)
	List(opts meta_v1.ListOptions) (*v1.PodVolumeRestoreList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PodVolumeRestore, err error)
	PodVolumeRestoreExpansion
}

// podVolumeRestores implements PodVolumeRestoreInterface
type podVolumeRestores struct {
	client rest.Interface
	ns     string
}

// newPodVolumeRestores returns a PodVolumeRestores
func newPodVolumeRestores(c *ArkV1Client, namespace string) *podVolumeRestores {
	return &podVolumeRestores{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the podVolumeRestore, and returns the corresponding podVolumeRestore object, and an error if there is any.
func (c *podVolumeRestores) Get(name string, options meta_v1.GetOptions) (result *v1.PodVolumeRestore, err error) {
	result = &v1.PodVolumeRestore{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podvolumerestores").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PodVolumeRestores that match those selectors.
func (c *podVolumeRestores) List(opts meta_v1.ListOptions) (result *v1.PodVolumeRestoreList, err error) {
	result = &v1.PodVolumeRestoreList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podvolumerestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested podVolumeRestores.
func (c *podVolumeRestores) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("podvolumerestores").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a podVolumeRestore and creates it.  Returns the server's representation of the podVolumeRestore, and an error, if there is any.
func (c *podVolumeRestores) Create(podVolumeRestore *v1.PodVolumeRestore) (result *v1.PodVolumeRestore, err error) {
	result = &v1.PodVolumeRestore{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("podvolumerestores").
		Body(podVolumeRestore).
		Do().
		Into(result)
	return
}

// Update takes the representation of a podVolumeRestore and updates it. Returns the server's representation of the podVolumeRestore, and an error, if there is any.
func (c *podVolumeRestores) Update(podVolumeRestore *v1.PodVolumeRestore) (result *v1.PodVolumeRestore, err error) {
	result = &v1.PodVolumeRestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("podvolumerestores").
		Name(podVolumeRestore.Name).
		Body(podVolumeRestore).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *podVolumeRestores) UpdateStatus(podVolumeRestore *v1.PodVolumeRestore) (result *v1.PodVolumeRestore, err error) {
	result = &v1.PodVolumeRestore{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("podvolumerestores").
		Name(podVolumeRestore.Name).
		SubResource("status").
		Body(podVolumeRestore).
		Do().
		Into(result)
	return
}

// Delete takes name of the podVolumeRestore and deletes it. Returns an error if one occurs.
func (c *podVolumeRestores) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podvolumerestores").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *podVolumeRestores) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podvolumerestores").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched podVolumeRestore.
func (c *podVolumeRestores) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PodVolumeRestore, err error) {
	result = &v1.PodVolumeRestore{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("podvolumerestores").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
