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


package source_impl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"

	yaml "github.com/ghodss/yaml"
	helm "helm.sh/helm/v3/pkg/release"
	app "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	podsecuritypolicy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    networking "k8s.io/api/networking/v1"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	cluster "containers-migration-factory/app/cluster"
	resource "containers-migration-factory/app/resource"
	MIGRATE_IMAGES "containers-migration-factory/controllers/MIGRATE_IMAGES"
)

//TODO: remove after all function migrated

var err error

var ignore_ds = map[string][]string{"kube-system": {"fluentd-gke", "gke-metrics-agent", "gke-metrics-agent-windows", "kube-proxy", "metadata-proxy-v0.1", "nvidia-gpu-device-plugin", "prometheus-to-sd"}}
var ignore_svc = map[string][]string{"kube-system": {"default-http-backend", "kube-dns", "metrics-server"}, "default": {"kubernetes"}}
var ignore_dep = map[string][]string{"kube-system": {"event-exporter-gke", "fluentd-gke-scaler", "kube-dns", "kube-dns-autoscaler", "l7-default-backend", "metrics-server-v0.3.6", "stackdriver-metadata-agent-cluster-level"}}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == a {
			return true
		}
	}
	return false
}

func itemExists(a []string, list []string) bool {
	for _, b := range list {
		for _, c := range a {
			if strings.ToLower(b) == strings.ToLower(c) {
				return true
			}
		}
	}
	return false
}

func trimCommonKeys(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == a {
			return true
		}
	}
	return false
}

func Trim_Item(ObjectMeta *metav1.ObjectMeta) {
	Trim_Item_All(ObjectMeta, true)
}

func Trim_Item_All(ObjectMeta *metav1.ObjectMeta, delGenAnnotation bool) {
	ObjectMeta.SelfLink = ""
	ObjectMeta.UID = ""
	ObjectMeta.ResourceVersion = ""
	ObjectMeta.Generation = 0
	ObjectMeta.CreationTimestamp = metav1.Time{}
	if delGenAnnotation {
		delete(ObjectMeta.Annotations, "deprecated.daemonset.template.generation")
	}
	delete(ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")

}

// Trim the unrequired fields from resource configuration
func Resource_trim_fields(resource_type string, resource *resource.Resources, resToInclude []string) {

	if resource_type == "Namespace" {
		var resource_list []v1.Namespace
		for _, item := range resource.Nsl.Items {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.Nsl.Items = resource_list
	}

	if resource_type == "MutatingWebhookConfiguration" && itemExists([]string{"MutatingWebhookConfigurations", "MutatingWebhookConfiguration", "all"}, resToInclude) {
		var resource_list []admissionregistration.MutatingWebhookConfiguration
		for _, item := range resource.MutatingWebhookConfigurationList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.MutatingWebhookConfigurationList = resource_list
	}

	if resource_type == "Ingress" && itemExists([]string{"Ingress", "Ingresses", "all"}, resToInclude) {
		var resource_list []networking.Ingress
		for _, item := range resource.IngressList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.IngressList = resource_list
	}

	if resource_type == "DaemonSet" && itemExists([]string{"daemonset", "daemonsets", "ds", "all"}, resToInclude) {
		var resource_list []app.DaemonSet
		for _, item := range resource.Dsl {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.Dsl = resource_list
	}

	if resource_type == "Service" && itemExists([]string{"service", "svc", "all"}, resToInclude) {
		var resource_list []v1.Service
		for _, item := range resource.Svcl {
			item.ObjectMeta.SelfLink = ""
			item.ObjectMeta.UID = ""
			item.ResourceVersion = ""
			item.Generation = 0
			item.CreationTimestamp = metav1.Time{}
			item.Spec.ClusterIP = ""
			item.Spec.ClusterIPs = nil
			for port, _ := range item.Spec.Ports {
				item.Spec.Ports[port].NodePort = 0
			}
			delete(item.ObjectMeta.Annotations, "deprecated.daemonset.template.generation")
			delete(item.ObjectMeta.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
			resource_list = append(resource_list, item)
		}
		resource.Svcl = resource_list
	}

	if resource_type == "Deployment" && itemExists([]string{"deployment", "deployments", "deploy", "all"}, resToInclude) {
		var resource_list []app.Deployment
		for _, item := range resource.Depl {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.Depl = resource_list
	}

	if resource_type == "Secrets" && itemExists([]string{"secrets", "secret", "all"}, resToInclude) {
		var resource_list []v1.Secret
		for _, item := range resource.SecretList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.SecretList = resource_list
	}

	if resource_type == "StorageClasses" && itemExists([]string{"storageclasses", "storageclass", "sc", "all"}, resToInclude) {
		var resource_list []storage.StorageClass
		for _, item := range resource.StorageClassList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.StorageClassList = resource_list
	}

	if resource_type == "ConfigMap" && itemExists([]string{"configmap", "configmaps", "cm", "all"}, resToInclude) {
		var resource_list []v1.ConfigMap
		for _, item := range resource.ConfigMapsList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.ConfigMapsList = resource_list
	}

	if resource_type == "Roles" && itemExists([]string{"role", "roles", "all"}, resToInclude) {
		var resource_list []rbac.Role
		for _, item := range resource.RoleList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.RoleList = resource_list
	}
	if resource_type == "PersistentVolumeClaim" && itemExists([]string{"persistentvolumeclaims", "persistentvolumeclaim", "pvc", "all"}, resToInclude) {
		var resource_list []v1.PersistentVolumeClaim
		for _, item := range resource.PersistentVolumeClaimsList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.PersistentVolumeClaimsList = resource_list
	}

	if resource_type == "CronJob" && itemExists([]string{"cronjobs", "cronjob", "cj", "all"}, resToInclude) {
		var resource_list []batchv1beta1.CronJob
		for _, item := range resource.CronJobList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.CronJobList = resource_list
	}

	if resource_type == "Job" && itemExists([]string{"jobs", "job", "all"}, resToInclude) {
		var resource_list []batchv1.Job
		for _, item := range resource.JobList {
			Trim_Item(&item.ObjectMeta)
			delete(item.Spec.Selector.MatchLabels,"controller-uid")
			delete(item.Spec.Template.Labels, "controller-uid")
			resource_list = append(resource_list, item)
		}
		resource.JobList = resource_list
	}

	if resource_type == "ValidatingWebhookConfiguration" && itemExists([]string{"validatingwebhookconfiguration", "validatingwebhookconfigurations", "all"}, resToInclude) {
		var resource_list []admissionregistration.ValidatingWebhookConfiguration
		for _, item := range resource.ValidatingWebhookConfigurationList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.ValidatingWebhookConfigurationList = resource_list
	}

	if resource_type == "RoleBindings" && itemExists([]string{"rolebinding", "rolebindings", "all"}, resToInclude) {
		var resource_list []rbac.RoleBinding
		for _, item := range resource.RoleBindingList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.RoleBindingList = resource_list
	}

	if resource_type == "ClusterRoles" && itemExists([]string{"clusterrole", "clusterroles", "all"}, resToInclude) {
		var resource_list []rbac.ClusterRole
		for _, item := range resource.ClusterRoleList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.ClusterRoleList = resource_list
	}

	if resource_type == "ClusterRoleBindings" && itemExists([]string{"clusterrolebinding", "clusterrolebindings", "all"}, resToInclude) {
		var resource_list []rbac.ClusterRoleBinding
		for _, item := range resource.ClusterRoleBindingList {
			Trim_Item(&item.ObjectMeta)
			resource_list = append(resource_list, item)
		}
		resource.ClusterRoleBindingList = resource_list
	}

	if resource_type == "HorizontalPodAutoscaler" && itemExists([]string{"horizontalpodautoscaler", "horizontalpodautoscalers"}, resToInclude) {
		var resource_list []autoscaling.HorizontalPodAutoscaler
		for _, item := range resource.HpaList {
			Trim_Item_All(&item.ObjectMeta, false)
			resource_list = append(resource_list, item)
		}
		resource.HpaList = resource_list
	}

	if resource_type == "PodSecurityPolicy" && itemExists([]string{"podsecuritypolicies", "podsecuritypolicy", "psp", "all"}, resToInclude) {
		var resource_list []podsecuritypolicy.PodSecurityPolicy
		for _, item := range resource.PspList {
			Trim_Item_All(&item.ObjectMeta, false)
			resource_list = append(resource_list, item)
		}
		resource.PspList = resource_list
	}

	if resource_type == "ServiceAccount" && itemExists([]string{"serviceaccount", "serviceaccounts", "sa", "all"}, resToInclude) {
		var resource_list []v1.ServiceAccount
		for _, item := range resource.SvcAccList {
			Trim_Item_All(&item.ObjectMeta, false)
			resource_list = append(resource_list, item)
		}
		resource.SvcAccList = resource_list
	}
}

// Scan source kubernetes cluster and generate the Job objects
func Generate_job_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("jobs", src.GetResources()) || stringInSlice("job", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			job, err := src.GetClientset().BatchV1().Jobs(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Secrets using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to glabal services list
			resource.JobList = append(resource.JobList, job.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the CronJob objects
func Generate_cronjob_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("cronjobs", src.GetResources()) || stringInSlice("cronjob", src.GetResources()) || stringInSlice("cj", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			cronjob, err := src.GetClientset().BatchV1beta1().CronJobs(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Secrets using cluster client: %v\n", err)
				os.Exit(1)
			}

            if src.Migrate_Images == "Yes" || src.Migrate_Images == "yes" {
                    for i, item := range cronjob.Items {
                            for j, image_spec := range item.Spec.JobTemplate.Spec.Template.Spec.Containers {
                                    image_name := image_spec.Image
                                    updated_image := MIGRATE_IMAGES.Validate(image_name, src.Registry_Names)
                                    if updated_image != "" {
                                            cronjob.Items[i].Spec.JobTemplate.Spec.Template.Spec.Containers[j].Image = updated_image
                                    }
                            }
                	}
			}

			// append list of services in this namespace to glabal services list
			resource.CronJobList = append(resource.CronJobList, cronjob.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the secret objects
func Generate_secret_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("secrets", src.GetResources()) || stringInSlice("secret", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			secret, err := src.GetClientset().CoreV1().Secrets(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Secrets using cluster client: %v\n", err)
				os.Exit(1)
			}

			// Remove default secret from each namespace
			for i, item := range secret.Items {
				//	if strings.Contains(item.ObjectMeta.Name, "default-token-") {
				if item.ObjectMeta.Annotations["kubernetes.io/service-account.name"] == "default" {
					secret.Items = append(secret.Items[:i], secret.Items[i+1:]...)
				}
			}

			// append list of services in this namespace to global services list
			resource.SecretList = append(resource.SecretList, secret.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the ConfigMap objects
func Generate_configmap_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("configmaps", src.GetResources()) || stringInSlice("configmap", src.GetResources()) || stringInSlice("cm", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			configmap, err := src.GetClientset().CoreV1().ConfigMaps(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes ConfigMaps using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.ConfigMapsList = append(resource.ConfigMapsList, configmap.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the MutatingWebhookConfiguration objects
func Generate_mutatingwebhook_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("mutatingWebhookconfigurations", src.GetResources()) || stringInSlice("mutatingwebhookconfiguration", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		mwc, err := src.GetClientset().AdmissionregistrationV1().MutatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes MutatingWebhookConfiguration using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of services in this namespace to glabal services list
		resource.MutatingWebhookConfigurationList = append(resource.MutatingWebhookConfigurationList, mwc.Items...)
	}
}

// Scan source kubernetes cluster and generate the ValidtingWebhookConfiguration objects
func Generate_validatingwebhook_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("validatingwebhookconfiguration", src.GetResources()) || stringInSlice("validatingwebhookconfigurations", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		vwc, err := src.GetClientset().AdmissionregistrationV1().ValidatingWebhookConfigurations().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes MutatingWebhookConfiguration using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of services in this namespace to glabal services list
		resource.ValidatingWebhookConfigurationList = append(resource.ValidatingWebhookConfigurationList, vwc.Items...)
	}
}

// Scan source kubernetes cluster and generate the ConfigMap objects
func Generate_ingress_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("ingresses", src.GetResources()) || stringInSlice("ingress", src.GetResources()) || stringInSlice("ing", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			ingress, err := src.GetClientset().NetworkingV1().Ingresses(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Ingresses using cluster client: %v\n", err)
				//os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.IngressList = append(resource.IngressList, ingress.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the Storage Class objects
func Generate_storage_class_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("storageclasses", src.GetResources()) || stringInSlice("storageclass", src.GetResources()) || stringInSlice("sc", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		//for _, element := range resource.Nsl.Items {
		sc, err := src.GetClientset().StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes Storage Classes using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of services in this namespace to global services list
		resource.StorageClassList = append(resource.StorageClassList, sc.Items...)
		//}
	}
}

// Scan source kubernetes cluster and generate the Persistent volume claim objects
func Generate_pvc_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("persistentvolumeclaims", src.GetResources()) || stringInSlice("persistentvolumeclaim", src.GetResources()) || stringInSlice("pvc", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			pvc, err := src.GetClientset().CoreV1().PersistentVolumeClaims(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Storage Classes using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to glabal services list
			resource.PersistentVolumeClaimsList = append(resource.PersistentVolumeClaimsList, pvc.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the service objects
func Generate_deployment_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("deployment", src.GetResources()) || stringInSlice("deployments", src.GetResources()) || stringInSlice("deploy", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			dep, err := src.GetClientset().AppsV1().Deployments(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes SVC using cluster client: %v\n", err)
				os.Exit(1)
			}
            
			if src.Migrate_Images == "Yes" || src.Migrate_Images == "yes" {
                    for i, item := range dep.Items {
                            for j, image_spec := range item.Spec.Template.Spec.Containers {
                                    image_name := image_spec.Image
                                    updated_image := MIGRATE_IMAGES.Validate(image_name, src.Registry_Names)
                                    if updated_image != "" {
                                            dep.Items[i].Spec.Template.Spec.Containers[j].Image = updated_image
                                    }
                            }
                	}
            }

			// append list of services in this namespace to global services list
			resource.Depl = append(resource.Depl, dep.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the service objects
func Generate_service_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("service", src.GetResources()) || stringInSlice("svc", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of services
		for _, element := range resource.Nsl.Items {
			svc, err := src.GetClientset().CoreV1().Services(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes SVC using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.Svcl = append(resource.Svcl, svc.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the daemonset objects
func Generate_daemonset_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("daemonset", src.GetResources()) || stringInSlice("daemonsets", src.GetResources()) || stringInSlice("ds", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of daemonsets
		for _, element := range resource.Nsl.Items {
			ds, err := src.GetClientset().AppsV1().DaemonSets(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Daemonsets using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.Dsl = append(resource.Dsl, ds.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the HPA objects
func Generate_hpa_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("horizontalpodautoscaler", src.GetResources()) || stringInSlice("horizontalpodautoscalers", src.GetResources()) || stringInSlice("hpa", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of hpas
		for _, element := range resource.Nsl.Items {
			hpa, err := src.GetClientset().AutoscalingV1().HorizontalPodAutoscalers(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			fmt.Println()
			if err != nil {
				fmt.Printf("Could not read kubernetes hpas using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to glabal services list
			resource.HpaList = append(resource.HpaList, hpa.Items...)
		}
	}
}

func Generate_psp_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("podsecuritypolicies", src.GetResources()) || stringInSlice("podsecuritypolicy", src.GetResources()) || stringInSlice("psp", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Get the list of pod security policies
		psp, err := src.GetClientset().PolicyV1beta1().PodSecurityPolicies().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes pod security policies using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of pod security policies to glabal services list
		resource.PspList = append(resource.PspList, psp.Items...)
	}
}

func Generate_serviceaccount_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("serviceaccount", src.GetResources()) || stringInSlice("serviceaccounts", src.GetResources()) || stringInSlice("sa", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of daemonsets
		for _, element := range resource.Nsl.Items {
			sa, err := src.GetClientset().CoreV1().ServiceAccounts(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes hpas using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of service accounts in this namespace to glabal services list
			resource.SvcAccList = append(resource.SvcAccList, sa.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the role objects
func Generate_role_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("role", src.GetResources()) || stringInSlice("roles", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of daemonsets
		for _, element := range resource.Nsl.Items {
			rl, err := src.GetClientset().RbacV1().Roles(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Role using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.RoleList = append(resource.RoleList, rl.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the role binding objects
func Generate_role_binding_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("rolebinding", src.GetResources()) || stringInSlice("rolebindings", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// Loop through all the namespaces and get the list of daemonsets
		for _, element := range resource.Nsl.Items {
			rbl, err := src.GetClientset().RbacV1().RoleBindings(element.ObjectMeta.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Could not read kubernetes Role Bindings using cluster client: %v\n", err)
				os.Exit(1)
			}

			// append list of services in this namespace to global services list
			resource.RoleBindingList = append(resource.RoleBindingList, rbl.Items...)
		}
	}
}

// Scan source kubernetes cluster and generate the cluster role objects
func Generate_cluster_role_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("clusterrole", src.GetResources()) || stringInSlice("clusterroles", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// not a namespaced resource and hence no loop through all the namespaces and get the list of clusterroles
		crl, err := src.GetClientset().RbacV1().ClusterRoles().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes Role Bindings using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of services in this namespace to global services list
		resource.ClusterRoleList = append(resource.ClusterRoleList, crl.Items...)
		// }
	}
}

// Scan source kubernetes cluster and generate the cluster role objects
func Generate_cluster_role_binding_config(src *cluster.Cluster, resource *resource.Resources) {
	if stringInSlice("clusterrolebinding", src.GetResources()) || stringInSlice("clusterrolebindings", src.GetResources()) || stringInSlice("all", src.GetResources()) {
		// not a namespaced resource and hence no loop through all the namespaces and get the list of clusterrole bindings
		crbl, err := src.GetClientset().RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not read kubernetes Role Bindings using cluster client: %v\n", err)
			os.Exit(1)
		}

		// append list of services in this namespace to global services list
		resource.ClusterRoleBindingList = append(resource.ClusterRoleBindingList, crbl.Items...)
		// }
	}
}

func Generate_namespace_list(src *cluster.Cluster, resource *resource.Resources) {
	if src.GetNamespaces() != nil && !(stringInSlice("all", src.GetNamespaces())) || (stringInSlice("all", src.GetNamespaces())) || (stringInSlice("namespaces", src.GetNamespaces())) || (stringInSlice("namespace", src.GetNamespaces())) || (stringInSlice("ns", src.GetNamespaces())) {
		// Intialize namespace list variable
		resource.Nsl = new(v1.NamespaceList)
		fmt.Println("Namespace list entered by user")

		//Loop through the list of namespace name entered. by used and get the namesapce object from cluster
		for _, element := range src.GetNamespaces() {
			if element == "kube-system" || element == "kube-public" || element == "kube-node-lease" {
				continue
			}
			ns, err := src.GetClientset().CoreV1().Namespaces().Get(context.TODO(), element, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Could not List kubernetes namespaces using cluster client: %v\n", err)
				os.Exit(1)
			}

			resource.Nsl.Items = append(resource.Nsl.Items, *ns)
		}
	} else {
		fmt.Println("Namespace list entered as 'all' by user, hence all namespaces will be considered")
		resource.Nsl, err = src.GetClientset().CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Could not List kubernetes namespaces using cluster client: %v\n", err)
			os.Exit(1)
		}

		j := 0
		for _, element := range resource.Nsl.Items {
			if element.ObjectMeta.Name != "kube-system" && element.ObjectMeta.Name != "kube-public" && element.ObjectMeta.Name != "kube-node-lease" {
				resource.Nsl.Items[j] = element
				j++
			}
		}
		resource.Nsl.Items = resource.Nsl.Items[:j]
	}

}

//Scan source kubernetes cluster and generate the Helm charts
func Generate_helm_charts(src *cluster.Cluster, resource *resource.Resources) {
	labelSelector := fmt.Sprintf("owner=helm")
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}

	resource.HelmList = make(map[string]map[string]string)
	for _, element := range resource.Nsl.Items {
		var helmCharts = make(map[string]helm.Release)

		secretList, err := src.GetClientset().CoreV1().Secrets(element.ObjectMeta.Name).List(context.TODO(), listOptions)
		if err != nil {
			fmt.Printf("Could not read kubernetes Secrets using cluster client: %v\n", err)
			os.Exit(1)
		}

		for _, secret := range secretList.Items {
			status_tmp := secret.ObjectMeta.Labels["status"]
			status := strings.Split(status_tmp, ":")
			if status[0] == "deployed" {

				base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(secret.Data["release"])))
				base64.StdEncoding.Decode(base64Text, []byte(secret.Data["release"]))

				var secret_uncompressed bytes.Buffer

				err = gunzipWrite(&secret_uncompressed, &base64Text)
				if err != nil {
					fmt.Println("test9")
					log.Fatal(err)
				}
				secret_string := secret_uncompressed.String()

				//convert the release string into helm release struct
				var secret_data helm.Release //map[string]interface{}
				json.Unmarshal([]byte(secret_string), &secret_data)

				//Add the chart name to the helmCharts global variable
				helmCharts[secret_data.Name] = secret_data
				path := src.Helm_path + "/KMFHelmCharts/namespaces/" + element.ObjectMeta.Name
				writeChartToFile(helmCharts, path, element.ObjectMeta.Name, resource)
			}

		}
	}
}

func writeChartToFile(charts map[string]helm.Release, path string, namespace string, resource *resource.Resources) {
	// Create the directory locally to store the helm charts
	fmt.Println("Path :", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0700)
		if err != nil {
			fmt.Println("Error creating the path for helm charts\n", err)
		}
	}

	var chartsPath = make(map[string]string)
	//var helmCharts = make(map[string]helm.Release)
	// Get the ctarts from release struct
	for k, v := range charts {

		// Create subdirectory to store the charts for this release
		if _, err := os.Stat(path + "/" + v.Name); os.IsNotExist(err) {
			err = os.Mkdir(path+"/"+v.Name, 0700)
			if err != nil {
				fmt.Println("Error creating the path for helm release: \n", err, v.Name)
			}
		}

		// Add path to the chart to the HelmList to later install the chart from this path on EKS cluster
		chartsPath[v.Name] = path + "/" + v.Name

		helm_templates := v.Chart.Templates
		fmt.Println("Chart Name:", k)
		//fmt.Println("template:", secret.Data["release"])
		for _, element := range helm_templates {

			fmt.Println("secrets:", element.Name)

			//fmt.Println("template:", string(element.Data))
			if _, err := os.Stat(filepath.Dir(path + "/" + v.Name + "/" + element.Name)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(path+"/"+v.Name+"/"+element.Name), 0700)
			}
			err := ioutil.WriteFile(path+"/"+v.Name+"/"+element.Name, element.Data, 0600)
			if err != nil {
				panic(err)
			}
		}

		helm_files := v.Chart.Files
		for _, element := range helm_files {
			fmt.Println("Files Name:", element.Name)
			//fmt.Println("template:", string(element.Data))
			if _, err := os.Stat(filepath.Dir(path + "/" + v.Name + "/" + element.Name)); os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(path+"/"+v.Name+"/"+element.Name), 0700)
			}
			err := ioutil.WriteFile(path+"/"+v.Name+"/"+element.Name, element.Data, 0600)
			if err != nil {
				panic(err)
			}
		}

		//Write values file
		if _, err := os.Stat(filepath.Dir(path + "/" + v.Name + "/" + "values.json")); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(path+"/"+v.Name+"/"+"values.json"), 0700)
		}

		jsonString, err := json.Marshal(v.Chart.Values)
		valuesyaml, err := yaml.JSONToYAML(jsonString)
		err = ioutil.WriteFile(path+"/"+v.Name+"/"+"values.yaml", valuesyaml, 0600)
		if err != nil {
			panic(err)
		}

		//Write Chart metadata to chart.json file
		if _, err := os.Stat(filepath.Dir(path + "/" + v.Name + "/" + "Chart.yaml")); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(path+"/"+v.Name+"/"+"Chart.yaml"), 0700)
		}

		jsonString, err = json.Marshal(v.Chart.Metadata)
		chartyaml, err := yaml.JSONToYAML(jsonString)
		err = ioutil.WriteFile(path+"/"+v.Name+"/"+"Chart.yaml", chartyaml, 0600)
		if err != nil {
			panic(err)
		}
	}

	resource.HelmList[namespace] = chartsPath
}

func gunzipWrite(w io.Writer, data *[]byte) error {
	// Write gzipped data to the client

	gr, err := gzip.NewReader(bytes.NewBuffer(*data))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer gr.Close()
	*data, err = ioutil.ReadAll(gr)

	w.Write(*data)
	return nil
}
