/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: MIT-0
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
 * permit persons to whom the Software is furnished to do so.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
 * PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
 * HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
 * OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package aks

import (
	cluster "containers-migration-factory/app/cluster"
	resource "containers-migration-factory/app/resource"
	"containers-migration-factory/app/source/source_impl"
	"fmt"
)

// AKS defines as source
type AKS struct{}

var log = false

func (c AKS) Connect(sCluster *cluster.Cluster) {
	sCluster.Generate_cluster_client()
}

func (g AKS) GetSourceDetails(sCluster *cluster.Cluster) resource.Resources {
	fmt.Println("AKS GetSourceDetails....")
	resources := resource.Resources{}
	source_impl.Generate_namespace_list(sCluster, &resources)

	source_impl.Generate_helm_charts(sCluster, &resources)

	source_impl.Generate_job_config(sCluster, &resources)

	source_impl.Generate_cronjob_config(sCluster, &resources)
	source_impl.Generate_secret_config(sCluster, &resources)
	source_impl.Generate_configmap_config(sCluster, &resources)
	//source_impl.Generate_mutatingwebhook_config(sCluster, &resources)
	//source_impl.Generate_validatingwebhook_config(sCluster, &resources)
	source_impl.Generate_ingress_config(sCluster, &resources)
	source_impl.Generate_storage_class_config(sCluster, &resources)
	source_impl.Generate_pvc_config(sCluster, &resources)
	source_impl.Generate_deployment_config(sCluster, &resources)

	source_impl.Generate_service_config(sCluster, &resources)
	source_impl.Generate_daemonset_config(sCluster, &resources)
	source_impl.Generate_hpa_config(sCluster, &resources)
	source_impl.Generate_psp_config(sCluster, &resources)
	source_impl.Generate_serviceaccount_config(sCluster, &resources)
	source_impl.Generate_role_config(sCluster, &resources)
	source_impl.Generate_role_binding_config(sCluster, &resources)
	source_impl.Generate_cluster_role_config(sCluster, &resources)
	source_impl.Generate_cluster_role_binding_config(sCluster, &resources)

	if log {
		fmt.Println("......JobList......", resources.JobList)
		fmt.Println("......Deployments......", resources.Depl)
		fmt.Println("......DaemonSet......", resources.Dsl)
		fmt.Println("......ServiceList......", resources.Svcl)
		fmt.Println("......StorageClassList......", resources.StorageClassList)
		fmt.Println("......ConfigMapsList......", resources.ConfigMapsList)
		fmt.Println("......IngressList......", resources.IngressList)
		fmt.Println("......RoleList......", resources.RoleList)
		fmt.Println("......RoleBindingList......", resources.RoleBindingList)
		fmt.Println("......ClusterRoleList......", resources.ClusterRoleList)
		fmt.Println("......ClusterRoleBindingList......", resources.ClusterRoleBindingList)
		fmt.Println("......HpaList......", resources.HpaList)
		fmt.Println("......PspList......", resources.PspList)
		fmt.Println("......SvcAccList......", resources.SvcAccList)
		fmt.Println("......CronJobList......", resources.CronJobList)
		fmt.Println("......PersistentVolumeClaimsList......", resources.PersistentVolumeClaimsList)
		//fmt.Println("......MutatingWebhookConfigurationList......", resources.MutatingWebhookConfigurationList)
		//fmt.Println("......ValidatingWebhookConfigurationList......", resources.ValidatingWebhookConfigurationList)
		fmt.Println("......HelmList......", resources.HelmList)
		fmt.Println("......JobList......", resources.JobList)
	}
	return resources
}

// AKS FormatSourceData implements the Geometry interface
func (g AKS) FormatSourceData(resource *resource.Resources, resToInclude []string) {
	fmt.Println("AKS FormatSourceData....start")
	source_impl.Resource_trim_fields("Namespace", resource, resToInclude)
	source_impl.Resource_trim_fields("DaemonSet", resource, resToInclude)
	//source_impl.Resource_trim_fields("MutatingWebhookConfiguration", resource, resToInclude)
	source_impl.Resource_trim_fields("Deployment", resource, resToInclude)
	source_impl.Resource_trim_fields("Service", resource, resToInclude)
	source_impl.Resource_trim_fields("Secrets", resource, resToInclude)
	source_impl.Resource_trim_fields("StorageClasses", resource, resToInclude)
	source_impl.Resource_trim_fields("Roles", resource, resToInclude)
	source_impl.Resource_trim_fields("RoleBindings", resource, resToInclude)
	source_impl.Resource_trim_fields("ClusterRoles", resource, resToInclude)
	source_impl.Resource_trim_fields("ClusterRoleBindings", resource, resToInclude)
	source_impl.Resource_trim_fields("HorizontalPodAutoscaler", resource, resToInclude)
	source_impl.Resource_trim_fields("PodSecurityPolicy", resource, resToInclude)
	source_impl.Resource_trim_fields("ServiceAccount", resource, resToInclude)
	source_impl.Resource_trim_fields("PersistentVolumeClaim", resource, resToInclude)
	source_impl.Resource_trim_fields("CronJob", resource, resToInclude)
	source_impl.Resource_trim_fields("Job", resource, resToInclude)
	source_impl.Resource_trim_fields("ConfigMap", resource, resToInclude)
	source_impl.Resource_trim_fields("Ingress", resource, resToInclude)
	fmt.Println("AKS FormatSourceData....End")
}
