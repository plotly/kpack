/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/pivotal/kpack/pkg/apis/experimental/v1alpha1"
	scheme "github.com/pivotal/kpack/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CustomClusterBuildersGetter has a method to return a CustomClusterBuilderInterface.
// A group's client should implement this interface.
type CustomClusterBuildersGetter interface {
	CustomClusterBuilders() CustomClusterBuilderInterface
}

// CustomClusterBuilderInterface has methods to work with CustomClusterBuilder resources.
type CustomClusterBuilderInterface interface {
	Create(*v1alpha1.CustomClusterBuilder) (*v1alpha1.CustomClusterBuilder, error)
	Update(*v1alpha1.CustomClusterBuilder) (*v1alpha1.CustomClusterBuilder, error)
	UpdateStatus(*v1alpha1.CustomClusterBuilder) (*v1alpha1.CustomClusterBuilder, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CustomClusterBuilder, error)
	List(opts v1.ListOptions) (*v1alpha1.CustomClusterBuilderList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CustomClusterBuilder, err error)
	CustomClusterBuilderExpansion
}

// customClusterBuilders implements CustomClusterBuilderInterface
type customClusterBuilders struct {
	client rest.Interface
}

// newCustomClusterBuilders returns a CustomClusterBuilders
func newCustomClusterBuilders(c *ExperimentalV1alpha1Client) *customClusterBuilders {
	return &customClusterBuilders{
		client: c.RESTClient(),
	}
}

// Get takes name of the customClusterBuilder, and returns the corresponding customClusterBuilder object, and an error if there is any.
func (c *customClusterBuilders) Get(name string, options v1.GetOptions) (result *v1alpha1.CustomClusterBuilder, err error) {
	result = &v1alpha1.CustomClusterBuilder{}
	err = c.client.Get().
		Resource("customclusterbuilders").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CustomClusterBuilders that match those selectors.
func (c *customClusterBuilders) List(opts v1.ListOptions) (result *v1alpha1.CustomClusterBuilderList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CustomClusterBuilderList{}
	err = c.client.Get().
		Resource("customclusterbuilders").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested customClusterBuilders.
func (c *customClusterBuilders) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("customclusterbuilders").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a customClusterBuilder and creates it.  Returns the server's representation of the customClusterBuilder, and an error, if there is any.
func (c *customClusterBuilders) Create(customClusterBuilder *v1alpha1.CustomClusterBuilder) (result *v1alpha1.CustomClusterBuilder, err error) {
	result = &v1alpha1.CustomClusterBuilder{}
	err = c.client.Post().
		Resource("customclusterbuilders").
		Body(customClusterBuilder).
		Do().
		Into(result)
	return
}

// Update takes the representation of a customClusterBuilder and updates it. Returns the server's representation of the customClusterBuilder, and an error, if there is any.
func (c *customClusterBuilders) Update(customClusterBuilder *v1alpha1.CustomClusterBuilder) (result *v1alpha1.CustomClusterBuilder, err error) {
	result = &v1alpha1.CustomClusterBuilder{}
	err = c.client.Put().
		Resource("customclusterbuilders").
		Name(customClusterBuilder.Name).
		Body(customClusterBuilder).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *customClusterBuilders) UpdateStatus(customClusterBuilder *v1alpha1.CustomClusterBuilder) (result *v1alpha1.CustomClusterBuilder, err error) {
	result = &v1alpha1.CustomClusterBuilder{}
	err = c.client.Put().
		Resource("customclusterbuilders").
		Name(customClusterBuilder.Name).
		SubResource("status").
		Body(customClusterBuilder).
		Do().
		Into(result)
	return
}

// Delete takes name of the customClusterBuilder and deletes it. Returns an error if one occurs.
func (c *customClusterBuilders) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("customclusterbuilders").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *customClusterBuilders) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("customclusterbuilders").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched customClusterBuilder.
func (c *customClusterBuilders) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CustomClusterBuilder, err error) {
	result = &v1alpha1.CustomClusterBuilder{}
	err = c.client.Patch(pt).
		Resource("customclusterbuilders").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}