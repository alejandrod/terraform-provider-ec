// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package enterprisesearchstate

import (
	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/terraform-providers/terraform-provider-ec/ec/ecresource/deploymentresource/deploymentstate"
)

// FlattenResources flattens Enterprise Search resources into its flattened structure.
func FlattenResources(in []*models.EnterpriseSearchResourceInfo, name string) []interface{} {
	var result = make([]interface{}, 0, len(in))
	for _, res := range in {
		var m = make(map[string]interface{})
		if isCurrentPlanEmpty(res) {
			continue
		}

		if res.Info.Name != nil && *res.Info.Name != name && *res.Info.Name != "" {
			m["display_name"] = *res.Info.Name
		}

		if res.RefID != nil && *res.RefID != "" {
			m["ref_id"] = *res.RefID
		}

		if res.Info.ID != nil && *res.Info.ID != "" {
			m["resource_id"] = *res.Info.ID
		}

		var plan = res.Info.PlanInfo.Current.Plan
		if plan.EnterpriseSearch != nil {
			m["version"] = plan.EnterpriseSearch.Version
		}

		if res.Region != nil {
			m["region"] = *res.Region
		}

		if topology := flattenTopology(plan); len(topology) > 0 {
			m["topology"] = topology
		}

		if res.ElasticsearchClusterRefID != nil {
			m["elasticsearch_cluster_ref_id"] = *res.ElasticsearchClusterRefID
		}

		if urls := deploymentstate.FlattenClusterEndpoint(res.Info.Metadata); len(urls) > 0 {
			for k, v := range urls {
				m[k] = v
			}
		}

		if c := flattenConfig(plan.EnterpriseSearch); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenTopology(plan *models.EnterpriseSearchPlan) []interface{} {
	var result = make([]interface{}, 0, len(plan.ClusterTopology))
	for _, topology := range plan.ClusterTopology {
		var m = make(map[string]interface{})
		if topology.Size == nil || topology.Size.Value == nil || *topology.Size.Value == 0 {
			continue
		}

		if topology.InstanceConfigurationID != "" {
			m["instance_configuration_id"] = topology.InstanceConfigurationID
		}

		if *topology.Size.Resource == "memory" {
			m["memory_per_node"] = deploymentstate.MemoryToState(*topology.Size.Value)
		}

		if nt := topology.NodeType; nt != nil {
			if nt.Appserver != nil {
				m["node_type_appserver"] = *nt.Appserver
			}

			if nt.Connector != nil {
				m["node_type_connector"] = *nt.Connector
			}

			if nt.Worker != nil {
				m["node_type_worker"] = *nt.Worker
			}
		}

		m["zone_count"] = topology.ZoneCount

		if c := flattenConfig(topology.EnterpriseSearch); len(c) > 0 {
			m["config"] = c
		}

		result = append(result, m)
	}

	return result
}

func flattenConfig(cfg *models.EnterpriseSearchConfiguration) []interface{} {
	var m = make(map[string]interface{})
	if cfg == nil {
		return nil
	}

	if cfg.UserSettingsYaml != "" {
		m["user_settings_yaml"] = cfg.UserSettingsYaml
	}

	if cfg.UserSettingsOverrideYaml != "" {
		m["user_settings_override_yaml"] = cfg.UserSettingsOverrideYaml
	}

	if cfg.UserSettingsJSON != nil {
		m["user_settings_json"] = cfg.UserSettingsJSON
	}

	if cfg.UserSettingsOverrideJSON != nil {
		m["user_settings_override_json"] = cfg.UserSettingsOverrideJSON
	}

	for k, v := range flattenSystemConfig(cfg.SystemSettings) {
		m[k] = v
	}

	if len(m) == 0 {
		return nil
	}

	return []interface{}{m}
}

func flattenSystemConfig(cfg *models.EnterpriseSearchSystemSettings) map[string]interface{} {
	var m = make(map[string]interface{})
	if cfg == nil {
		return nil
	}

	if cfg.SecretSessionKey != "" {
		m["secret_session_key"] = cfg.SecretSessionKey
	}

	if len(m) == 0 {
		return nil
	}

	return m
}

func isCurrentPlanEmpty(res *models.EnterpriseSearchResourceInfo) bool {
	var emptyPlanInfo = res.Info == nil || res.Info.PlanInfo == nil || res.Info.PlanInfo.Current == nil
	return emptyPlanInfo || res.Info.PlanInfo.Current.Plan == nil
}
