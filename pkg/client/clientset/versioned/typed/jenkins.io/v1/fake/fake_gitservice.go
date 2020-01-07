// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	jenkinsiov1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGitServices implements GitServiceInterface
type FakeGitServices struct {
	Fake *FakeJenkinsV1
	ns   string
}

var gitservicesResource = schema.GroupVersionResource{Group: "jenkins.io", Version: "v1", Resource: "gitservices"}

var gitservicesKind = schema.GroupVersionKind{Group: "jenkins.io", Version: "v1", Kind: "GitService"}

// Get takes name of the gitService, and returns the corresponding gitService object, and an error if there is any.
func (c *FakeGitServices) Get(name string, options v1.GetOptions) (result *jenkinsiov1.GitService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(gitservicesResource, c.ns, name), &jenkinsiov1.GitService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.GitService), err
}

// List takes label and field selectors, and returns the list of GitServices that match those selectors.
func (c *FakeGitServices) List(opts v1.ListOptions) (result *jenkinsiov1.GitServiceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(gitservicesResource, gitservicesKind, c.ns, opts), &jenkinsiov1.GitServiceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &jenkinsiov1.GitServiceList{ListMeta: obj.(*jenkinsiov1.GitServiceList).ListMeta}
	for _, item := range obj.(*jenkinsiov1.GitServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested gitServices.
func (c *FakeGitServices) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(gitservicesResource, c.ns, opts))

}

// Create takes the representation of a gitService and creates it.  Returns the server's representation of the gitService, and an error, if there is any.
func (c *FakeGitServices) Create(gitService *jenkinsiov1.GitService) (result *jenkinsiov1.GitService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(gitservicesResource, c.ns, gitService), &jenkinsiov1.GitService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.GitService), err
}

// Update takes the representation of a gitService and updates it. Returns the server's representation of the gitService, and an error, if there is any.
func (c *FakeGitServices) Update(gitService *jenkinsiov1.GitService) (result *jenkinsiov1.GitService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(gitservicesResource, c.ns, gitService), &jenkinsiov1.GitService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.GitService), err
}

// Delete takes name of the gitService and deletes it. Returns an error if one occurs.
func (c *FakeGitServices) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(gitservicesResource, c.ns, name), &jenkinsiov1.GitService{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGitServices) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(gitservicesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &jenkinsiov1.GitServiceList{})
	return err
}

// Patch applies the patch and returns the patched gitService.
func (c *FakeGitServices) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *jenkinsiov1.GitService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(gitservicesResource, c.ns, name, pt, data, subresources...), &jenkinsiov1.GitService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*jenkinsiov1.GitService), err
}
