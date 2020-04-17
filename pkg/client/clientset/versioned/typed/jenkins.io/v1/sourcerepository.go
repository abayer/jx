// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/jenkins-x/jx/v2/pkg/apis/jenkins.io/v1"
	scheme "github.com/jenkins-x/jx/v2/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SourceRepositoriesGetter has a method to return a SourceRepositoryInterface.
// A group's client should implement this interface.
type SourceRepositoriesGetter interface {
	SourceRepositories(namespace string) SourceRepositoryInterface
}

// SourceRepositoryInterface has methods to work with SourceRepository resources.
type SourceRepositoryInterface interface {
	Create(*v1.SourceRepository) (*v1.SourceRepository, error)
	Update(*v1.SourceRepository) (*v1.SourceRepository, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.SourceRepository, error)
	List(opts metav1.ListOptions) (*v1.SourceRepositoryList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SourceRepository, err error)
	SourceRepositoryExpansion
}

// sourceRepositories implements SourceRepositoryInterface
type sourceRepositories struct {
	client rest.Interface
	ns     string
}

// newSourceRepositories returns a SourceRepositories
func newSourceRepositories(c *JenkinsV1Client, namespace string) *sourceRepositories {
	return &sourceRepositories{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sourceRepository, and returns the corresponding sourceRepository object, and an error if there is any.
func (c *sourceRepositories) Get(name string, options metav1.GetOptions) (result *v1.SourceRepository, err error) {
	result = &v1.SourceRepository{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sourcerepositories").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SourceRepositories that match those selectors.
func (c *sourceRepositories) List(opts metav1.ListOptions) (result *v1.SourceRepositoryList, err error) {
	result = &v1.SourceRepositoryList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sourcerepositories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sourceRepositories.
func (c *sourceRepositories) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sourcerepositories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a sourceRepository and creates it.  Returns the server's representation of the sourceRepository, and an error, if there is any.
func (c *sourceRepositories) Create(sourceRepository *v1.SourceRepository) (result *v1.SourceRepository, err error) {
	result = &v1.SourceRepository{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sourcerepositories").
		Body(sourceRepository).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sourceRepository and updates it. Returns the server's representation of the sourceRepository, and an error, if there is any.
func (c *sourceRepositories) Update(sourceRepository *v1.SourceRepository) (result *v1.SourceRepository, err error) {
	result = &v1.SourceRepository{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sourcerepositories").
		Name(sourceRepository.Name).
		Body(sourceRepository).
		Do().
		Into(result)
	return
}

// Delete takes name of the sourceRepository and deletes it. Returns an error if one occurs.
func (c *sourceRepositories) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sourcerepositories").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sourceRepositories) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sourcerepositories").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sourceRepository.
func (c *sourceRepositories) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.SourceRepository, err error) {
	result = &v1.SourceRepository{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sourcerepositories").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
