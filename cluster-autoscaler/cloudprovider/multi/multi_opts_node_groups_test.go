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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
)

func TestNodeGroupsAndNodeGroupSpecs(t *testing.T) {
	opts := &coreoptions.AutoscalerOptions{AutoscalingOptions: config.AutoscalingOptions{NodeGroups: []string{
		"gce:0:2:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/cert",
		"aws:0:1:aws-cert",
		"aws-eu-west1:0:2:aws-gpu",
		"gce:0:3:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/gpu",
	}}}
	do := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: []string{
		"gce:0:2:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/cert",
		"aws:0:1:aws-cert",
		"aws-eu-west1:0:2:aws-gpu",
		"gce:0:3:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/gpu",
	}}

	providerOpts, providerDiscoveryOptions, err := GetNodeGroupsAndNodeGroupSpecs(opts, do, provider{"gce", "cloud_config_gce"})

	assert.Nil(t, err)
	assert.Equal(t, providerOpts.CloudProviderName, "gce")
	assert.Equal(t, providerOpts.CloudConfig, "cloud_config_gce")
	assert.Equal(t, len(providerOpts.NodeGroups), 2)
	assert.Equal(t, providerOpts.NodeGroups[0], "0:2:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/cert")
	assert.Equal(t, providerOpts.NodeGroups[1], "0:3:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/gpu")
	assert.Equal(t, len(providerDiscoveryOptions.NodeGroupSpecs), 2)
	assert.Equal(t, providerDiscoveryOptions.NodeGroupSpecs[0], "0:2:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/cert")
	assert.Equal(t, providerDiscoveryOptions.NodeGroupSpecs[1], "0:3:https://www.googleapis.com/compute/v1/projects/my-project-123456/zones/europe-west1-b/instanceGroups/gpu")

	providerOpts, providerDiscoveryOptions, err = GetNodeGroupsAndNodeGroupSpecs(opts, do, provider{"aws", "cloud_config_aws"})

	assert.Nil(t, err)
	assert.Equal(t, providerOpts.CloudProviderName, "aws")
	assert.Equal(t, providerOpts.CloudConfig, "cloud_config_aws")
	assert.Equal(t, len(providerOpts.NodeGroups), 1)
	assert.Equal(t, providerOpts.NodeGroups[0], "0:1:aws-cert")
	assert.Equal(t, len(providerDiscoveryOptions.NodeGroupSpecs), 1)
	assert.Equal(t, providerDiscoveryOptions.NodeGroupSpecs[0], "0:1:aws-cert")

	providerOpts, providerDiscoveryOptions, err = GetNodeGroupsAndNodeGroupSpecs(opts, do, provider{"aws-eu-west1", "cloud_config_aws"})

	assert.Nil(t, err)
	assert.Equal(t, providerOpts.CloudProviderName, "aws-eu-west1")
	assert.Equal(t, providerOpts.CloudConfig, "cloud_config_aws")
	assert.Equal(t, len(providerOpts.NodeGroups), 1)
	assert.Equal(t, providerOpts.NodeGroups[0], "0:2:aws-gpu")
	assert.Equal(t, len(providerDiscoveryOptions.NodeGroupSpecs), 1)
	assert.Equal(t, providerDiscoveryOptions.NodeGroupSpecs[0], "0:2:aws-gpu")
}
