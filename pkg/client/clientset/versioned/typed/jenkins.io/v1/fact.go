// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	scheme "github.com/jenkins-x/jx/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// FactsGetter has a method to return a FactInterface.
// A group's client should implement this interface.
type FactsGetter interface {
	Facts(namespace string) FactInterface
}

// FactInterface has methods to work with Fact resources.
type FactInterface interface {
	Create(*v1.Fact) (*v1.Fact, error)
	Update(*v1.Fact) (*v1.Fact, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Fact, error)
	List(opts metav1.ListOptions) (*v1.FactList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Fact, err error)
	FactExpansion
}

// facts implements FactInterface
type facts struct {
	client rest.Interface
	ns     string
}

// newFacts returns a Facts
func newFacts(c *JenkinsV1Client, namespace string) *facts {
	return &facts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the fact, and returns the corresponding fact object, and an error if there is any.
func (c *facts) Get(name string, options metav1.GetOptions) (result *v1.Fact, err error) {
	result = &v1.Fact{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("facts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Facts that match those selectors.
func (c *facts) List(opts metav1.ListOptions) (result *v1.FactList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.FactList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("facts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested facts.
func (c *facts) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("facts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a fact and creates it.  Returns the server's representation of the fact, and an error, if there is any.
func (c *facts) Create(fact *v1.Fact) (result *v1.Fact, err error) {
	result = &v1.Fact{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("facts").
		Body(fact).
		Do().
		Into(result)
	return
}

// Update takes the representation of a fact and updates it. Returns the server's representation of the fact, and an error, if there is any.
func (c *facts) Update(fact *v1.Fact) (result *v1.Fact, err error) {
	result = &v1.Fact{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("facts").
		Name(fact.Name).
		Body(fact).
		Do().
		Into(result)
	return
}

// Delete takes name of the fact and deletes it. Returns an error if one occurs.
func (c *facts) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("facts").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *facts) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("facts").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched fact.
func (c *facts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Fact, err error) {
	result = &v1.Fact{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("facts").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
