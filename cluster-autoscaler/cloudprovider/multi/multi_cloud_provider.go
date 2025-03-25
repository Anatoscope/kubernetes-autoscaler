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
	"fmt"
	"maps"
	_ "unsafe"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	autoscalererrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "multi/accelerator"
)

// MultiCloudProvider implements CloudProvider interface.
type MultiCloudProvider struct {
	providers                []cloudprovider.CloudProvider
	resourceLimiterFromFlags *cloudprovider.ResourceLimiter
}

//go:linkname buildCloudProvider k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder.buildCloudProvider
func buildCloudProvider(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider

// Cleanup cleans up all resources before the cloud provider is removed
func (multi *MultiCloudProvider) Cleanup() error {
	var allErr error = nil
	for _, provider := range multi.providers {
		errors.Join(allErr, provider.Cleanup())
	}
	return allErr
}

// Name returns name of the cloud provider.
func (multi *MultiCloudProvider) Name() string {
	return cloudprovider.MultiProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (multi *MultiCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (multi *MultiCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	availableGPUTypes := make(map[string]struct{})
	for _, provider := range multi.providers {
		maps.Copy(availableGPUTypes, provider.GetAvailableGPUTypes())
	}
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (multi *MultiCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(multi, node)
}

// NodeGroups returns all node groups configured for this cloud provider.
func (multi *MultiCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	var nodeGroups []cloudprovider.NodeGroup
	for _, provider := range multi.providers {
		nodeGroups = append(nodeGroups, provider.NodeGroups()...)
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node.
func (multi *MultiCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	for _, provider := range multi.providers {
		if nodeGroup, err := provider.NodeGroupForNode(node); err == nil && nodeGroup != nil {
			return nodeGroup, nil
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (multi *MultiCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	hasInstance := false
	var allErr error = nil
	for _, provider := range multi.providers {
		providerHasInstance, err := provider.HasInstance(node)
		if err == nil {
			hasInstance = hasInstance || providerHasInstance
		}
		allErr = errors.Join(allErr, err)
	}
	return hasInstance, allErr
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (multi *MultiCloudProvider) Pricing() (cloudprovider.PricingModel, autoscalererrors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (multi *MultiCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	var machineTypes []string
	var allErr error = nil
	for _, provider := range multi.providers {
		providerMachineTypes, err := provider.GetAvailableMachineTypes()
		if err == nil {
			machineTypes = append(machineTypes, providerMachineTypes...)
		}
		errors.Join(allErr, err)
	}
	return machineTypes, allErr
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (multi *MultiCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	// TODO: if not implemented in any then remove NewNodeGroup() from all providers and core-provider interface
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (multi *MultiCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	// TODO: if not implemented in any then remove ResourceLimiter from all providers and core-provider interface
	return multi.resourceLimiterFromFlags, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (multi *MultiCloudProvider) Refresh() error {
	var allErr error = nil
	for _, provider := range multi.providers {
		errors.Join(allErr, provider.Refresh())
	}
	return allErr
}

// BuildMulti builds the multi-cloud provider.
func BuildMulti(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	cloudConfig, err := newConfig(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("failed to create multi cloud provider: %v", fmt.Errorf("unable to create cloud config: %w", err))
		return nil
	}

	var providers []cloudprovider.CloudProvider

	for _, provider := range cloudConfig.Providers {
		var memberOpts *coreoptions.AutoscalerOptions
		var discoveryOptions cloudprovider.NodeGroupDiscoveryOptions
		memberOpts, discoveryOptions, err = GetNodeGroupsAndNodeGroupSpecs(opts, do, provider)
		if err != nil {
			klog.Fatalf("failed to create multi cloud provider: %v", err)
			return nil
		}
		providers = append(providers, buildCloudProvider(memberOpts, discoveryOptions, rl))
	}

	return &MultiCloudProvider{
		providers:                providers,
		resourceLimiterFromFlags: rl,
	}
}
