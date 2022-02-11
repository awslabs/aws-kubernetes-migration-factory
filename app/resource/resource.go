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
 

package resource

import (
	
	app "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	podsecuritypolicy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	storage "k8s.io/api/storage/v1"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Resources struct {
	Svcl       []v1.Service
	Nsl        *v1.NamespaceList
	Dsl        []app.DaemonSet
	SecretList []v1.Secret
	//var map[string]
	Depl                               []app.Deployment
	StorageClassList                   []storage.StorageClass
	ConfigMapsList                     []v1.ConfigMap
	IngressList                        []networking.Ingress
	RoleList                           []rbac.Role
	RoleBindingList                    []rbac.RoleBinding
	ClusterRoleList                    []rbac.ClusterRole
	ClusterRoleBindingList             []rbac.ClusterRoleBinding
	HpaList                            []autoscaling.HorizontalPodAutoscaler
	PspList                            []podsecuritypolicy.PodSecurityPolicy
	SvcAccList                         []v1.ServiceAccount
	CronJobList                        []batchv1beta1.CronJob
	JobList                            []batchv1.Job
	PersistentVolumeClaimsList         []v1.PersistentVolumeClaim
	MutatingWebhookConfigurationList   []admissionregistration.MutatingWebhookConfiguration
	ValidatingWebhookConfigurationList []admissionregistration.ValidatingWebhookConfiguration
	HelmList						    map[string]map[string]string // Helm data namespace: [ release name : path to chart]
	//	crdList *unstructured.UnstructuredList //[]apiextensions.CustomResourceDefinition
}