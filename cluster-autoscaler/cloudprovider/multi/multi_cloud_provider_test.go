/*
Copyright 2023 The Kubernetes Authors.

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

package multi

import (
	"errors"
	"testing"

	mockprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/mocks"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCloudProvider is a mock implementation of the CloudProvider interface.
type MockCloudProvider struct {
	mock.Mock
}

var node apiv1.Node = apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}

// No providers
func TestMultiCloudProviderHasInstanceNoProviders(t *testing.T) {
	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{}}
	hasInstance, err := multi.HasInstance(&node)
	assert.False(t, hasInstance)
	assert.Nil(t, err)
}

// One provider returns true, no error
func TestMultiCloudProviderHasInstanceOne(t *testing.T) {
	provider1 := &mockprovider.CloudProvider{}
	provider1.On("HasInstance", &node).Return(true, nil)

	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{provider1}}
	hasInstance, err := multi.HasInstance(&node)
	assert.True(t, hasInstance)
	assert.Nil(t, err)
}

// Multiple providers, one returns true, another false
func TestMultiCloudProviderHasInstanceTwo(t *testing.T) {
	provider1 := &mockprovider.CloudProvider{}
	provider1.On("HasInstance", &node).Return(true, nil)

	provider2 := &mockprovider.CloudProvider{}
	provider2.On("HasInstance", &node).Return(false, nil)

	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{provider1, provider2}}
	hasInstance, err := multi.HasInstance(&node)
	assert.True(t, hasInstance)
	assert.Nil(t, err)
}

// Multiple providers, all return false
func TestMultiCloudProviderHasInstanceTwoFalse(t *testing.T) {
	provider1 := &mockprovider.CloudProvider{}
	provider1.On("HasInstance", &node).Return(false, nil)

	provider2 := &mockprovider.CloudProvider{}
	provider2.On("HasInstance", &node).Return(false, nil)

	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{provider1, provider2}}
	hasInstance, err := multi.HasInstance(&node)
	assert.False(t, hasInstance)
	assert.Nil(t, err)
}

// One provider returns an error
func TestMultiCloudProviderHasInstanceTwoWithError(t *testing.T) {
	provider1 := &mockprovider.CloudProvider{}
	provider1.On("HasInstance", &node).Return(false, nil)

	provider2 := &mockprovider.CloudProvider{}
	provider2.On("HasInstance", &node).Return(false, errors.New("API error"))

	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{provider1, provider2}}
	hasInstance, err := multi.HasInstance(&node)
	assert.False(t, hasInstance)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API error")
}

// All providers return errors
func TestMultiCloudProviderHasInstanceAllErrors(t *testing.T) {
	provider1 := &mockprovider.CloudProvider{}
	provider1.On("HasInstance", &node).Return(false, errors.New("API error"))

	provider2 := &mockprovider.CloudProvider{}
	provider2.On("HasInstance", &node).Return(false, errors.New("API error"))

	multi := &MultiCloudProvider{providers: []cloudprovider.CloudProvider{provider1, provider2}}
	hasInstance, err := multi.HasInstance(&node)
	assert.False(t, hasInstance)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "API error")
}
