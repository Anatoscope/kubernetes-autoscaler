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
	"fmt"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
)

func extractNodeGroupOptionsForProvider(allNodeGroupOptions []string, providerName string) ([]string, error) {
	// Options are prefixed with "<providername>:".

	nodeGroupOptions := make([]string, 0)
	for _, providerNodeGroupOption := range allNodeGroupOptions {
		p, nodeGroupOption, found := strings.Cut(providerNodeGroupOption, ":")
		if !found {
			return []string{}, fmt.Errorf("wrong nodes configuration: %s", nodeGroupOption)
		}

		if p == providerName {
			nodeGroupOptions = append(nodeGroupOptions, nodeGroupOption)
		}
	}

	return nodeGroupOptions, nil
}

func GetNodeGroupsAndNodeGroupSpecs(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, provider provider) (*coreoptions.AutoscalerOptions, cloudprovider.NodeGroupDiscoveryOptions, error) {
	providerOpts := *opts
	providerDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{}

	var err error

	providerOpts.NodeGroups, err = extractNodeGroupOptionsForProvider(opts.NodeGroups, provider.Name)
	if err != nil {
		return nil, cloudprovider.NodeGroupDiscoveryOptions{}, err
	}

	providerDiscoveryOptions.NodeGroupSpecs, err = extractNodeGroupOptionsForProvider(do.NodeGroupSpecs, provider.Name)
	if err != nil {
		return nil, cloudprovider.NodeGroupDiscoveryOptions{}, err
	}

	providerDiscoveryOptions.NodeGroupAutoDiscoverySpecs, err = extractNodeGroupOptionsForProvider(do.NodeGroupAutoDiscoverySpecs, provider.Name)
	if err != nil {
		return nil, cloudprovider.NodeGroupDiscoveryOptions{}, err
	}

	providerOpts.CloudProviderName = provider.Name
	providerOpts.CloudConfig = provider.CloudConfig

	return &providerOpts, providerDiscoveryOptions, nil
}
