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


package AWS

import (
	"context"
	"fmt"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/strvals"
	cluster "containers-migration-factory/app/cluster"
	resource "containers-migration-factory/app/resource"
)


type Config struct {
	CurrentContext string `yaml:"current-context"`
}

var config Config
var settings *cli.EnvSettings

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == a {
			return true
		}
	}
	return false
}

func Deploy_resource_eks(dst *cluster.Cluster, src_resources *resource.Resources) {

	// Create non-namespaces resources
	if stringInSlice("mutatingWebhookconfigurations", dst.Resources) || stringInSlice("mutatingwebhookconfiguration", dst.Resources) || stringInSlice("all", dst.Resources) {
		// Create list of MutatingWebhookCOnfiguration in destination cluster
		for _, element := range src_resources.MutatingWebhookConfigurationList {
			element := element
			fmt.Println("Creating MutatingWebhook: ", element.ObjectMeta.Name)
			_, err := dst.Clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), &element, metav1.CreateOptions{})
			if err != nil {
				fmt.Println(err)
			}

		}
	}

	// Create list of ValidatingWebhookCOnfiguration in destination cluster
	if stringInSlice("validatingwebhookconfiguration", dst.Resources) || stringInSlice("validatingwebhookconfigurations", dst.Resources) || stringInSlice("all", dst.Resources) {
		for _, element := range src_resources.ValidatingWebhookConfigurationList {
			element := element
			fmt.Println("Creating MutatingWebhook: ", element.ObjectMeta.Name)
			_, err := dst.Clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().Create(context.TODO(), &element, metav1.CreateOptions{})
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	// Create list of namespaces in destination cluster
	for _, element := range src_resources.Nsl.Items {
		element := element
		fmt.Println("Creating the namespace: ", element.ObjectMeta.Name)
		_, err := dst.Clientset.CoreV1().Namespaces().Create(context.TODO(), &element, metav1.CreateOptions{})
		if err != nil {
			fmt.Println(err)
		}
	}

	// Install/Upgrade helm charts 
	Deploy_helm_charts(dst, src_resources)

	// Loop through each namespace and create resources inside each namespace
	for _, element := range src_resources.Nsl.Items {
		element := element

		fmt.Println("=====================================================================")
		fmt.Println("Operating on namespace: ", element.ObjectMeta.Name)
		fmt.Println("=====================================================================")

		// Create secrets resource
		if stringInSlice("secrets", dst.Resources) || stringInSlice("secret", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Secrets")
			for _, secret := range src_resources.SecretList {
				secret := secret
				if secret.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Secret: ", secret.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().Secrets(element.ObjectMeta.Name).Create(context.TODO(), &secret, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create ConfigMap resource
		if stringInSlice("configmaps", dst.Resources) || stringInSlice("configmap", dst.Resources) || stringInSlice("cm", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating ConfigMap's")
			for _, cm := range src_resources.ConfigMapsList {
				cm := cm
				if cm.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating ConfigMap: ", cm.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().ConfigMaps(element.ObjectMeta.Name).Create(context.TODO(), &cm, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create StorageClass resource
		if stringInSlice("storageclasses", dst.Resources) || stringInSlice("storageclass", dst.Resources) || stringInSlice("sc", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating StorageClasses")
			for _, sc := range src_resources.StorageClassList {
				sc := sc
				if sc.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating StorageClass: ", sc.ObjectMeta.Name)
					_, err := dst.Clientset.StorageV1().StorageClasses().Create(context.TODO(), &sc, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create PVC resource
		if stringInSlice("persistentvolumeclaims", dst.Resources) || stringInSlice("persistentvolumeclaim", dst.Resources) || stringInSlice("pvc", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating PersistentVolumeClaims")
			for _, pvc := range src_resources.PersistentVolumeClaimsList {
				pvc := pvc
				if pvc.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating PVC: ", pvc.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().PersistentVolumeClaims(element.ObjectMeta.Name).Create(context.TODO(), &pvc, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create deployment resource
		if stringInSlice("deployment", dst.Resources) || stringInSlice("deployments", dst.Resources) || stringInSlice("deploy", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Deployment")
			for _, dep := range src_resources.Depl {
				dep := dep
				if dep.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Deployment: ", dep.ObjectMeta.Name)
					_, err := dst.Clientset.AppsV1().Deployments(element.ObjectMeta.Name).Create(context.TODO(), &dep, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Service resource
		if stringInSlice("service", dst.Resources) || stringInSlice("svc", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Service")
			for _, svc := range src_resources.Svcl {
				svc := svc
				if svc.ObjectMeta.Namespace == element.ObjectMeta.Name {
					svc.Spec.ClusterIP = ""
					for port, _ := range svc.Spec.Ports {
						svc.Spec.Ports[port].NodePort = 0
					}
					fmt.Println("Creating Service: ", svc.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().Services(element.ObjectMeta.Name).Create(context.TODO(), &svc, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Daemonset resource
		if stringInSlice("daemonset", dst.Resources) || stringInSlice("daemonsets", dst.Resources) || stringInSlice("ds", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Daemonset")
			for _, ds := range src_resources.Dsl {
				ds := ds
				if ds.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Daemonset: ", ds.ObjectMeta.Name)
					_, err := dst.Clientset.AppsV1().DaemonSets(element.ObjectMeta.Name).Create(context.TODO(), &ds, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Ingress resource
		if stringInSlice("ingresses", dst.Resources) || stringInSlice("ingress", dst.Resources) || stringInSlice("ing", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Ingresses")
			for _, ingress := range src_resources.IngressList {
				ingress := ingress
				if ingress.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Ingress: ", ingress.ObjectMeta.Name)
					_, err := dst.Clientset.NetworkingV1().Ingresses(element.ObjectMeta.Name).Create(context.TODO(), &ingress, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Roles resource
		if stringInSlice("role", dst.Resources) || stringInSlice("roles", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Roles")
			for _, rl := range src_resources.RoleList {
				rl := rl
				if rl.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Roles: ", rl.ObjectMeta.Name)
					_, err := dst.Clientset.RbacV1().Roles(element.ObjectMeta.Name).Create(context.TODO(), &rl, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Role Bindings resource
		if stringInSlice("rolebinding", dst.Resources) || stringInSlice("rolebindings", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Role Bindings")
			for _, rbl := range src_resources.RoleBindingList {
				rbl := rbl
				if rbl.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Role Bindings: ", rbl.ObjectMeta.Name)
					_, err := dst.Clientset.RbacV1().RoleBindings(element.ObjectMeta.Name).Create(context.TODO(), &rbl, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create secrets resource
		if stringInSlice("jobs", dst.Resources) || stringInSlice("job", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Secrets")
			for _, secret := range src_resources.SecretList {
				secret := secret
				if secret.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Secret: ", secret.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().Secrets(element.ObjectMeta.Name).Create(context.TODO(), &secret, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create StorageClass resource
		if stringInSlice("storageclasses", dst.Resources) || stringInSlice("storageclass", dst.Resources) || stringInSlice("sc", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating StorageClasses")
			for _, sc := range src_resources.StorageClassList {
				sc := sc
				if sc.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating StorageClass: ", sc.ObjectMeta.Name)
					_, err := dst.Clientset.StorageV1().StorageClasses().Create(context.TODO(), &sc, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create CronJob resource
		if stringInSlice("cronjobs", dst.Resources) || stringInSlice("cronjob", dst.Resources) || stringInSlice("cj", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating CronJob's")
			for _, cronjob := range src_resources.CronJobList {		
				cronjob := cronjob		
				if cronjob.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating CronJob: ", cronjob.ObjectMeta.Name)
					_, err := dst.Clientset.BatchV1beta1().CronJobs(element.ObjectMeta.Name).Create(context.TODO(), &cronjob, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create job resource
		if stringInSlice("job", dst.Resources) || stringInSlice("jobs", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Job's")
			for _, job := range src_resources.JobList {			
				job := job	
				if job.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Job's: ", job.ObjectMeta.Name)
					_, err := dst.Clientset.BatchV1().Jobs(element.ObjectMeta.Name).Create(context.TODO(), &job, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Cluster Cluster Role resource
		if stringInSlice("clusterrole", dst.Resources) || stringInSlice("clusterroles", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Cluster Roles")
			for _, crl := range src_resources.ClusterRoleList {
				crl := crl			
				if crl.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Role Bindings: ", crl.ObjectMeta.Name)
					_, err := dst.Clientset.RbacV1().ClusterRoles().Create(context.TODO(), &crl, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Cluster Role Binding resource
		if stringInSlice("clusterrolebinding", dst.Resources) || stringInSlice("clusterrolebindings", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Cluster Role Bindings")
			for _, crbl := range src_resources.ClusterRoleBindingList {
				crbl := crbl
				if crbl.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Role Bindings: ", crbl.ObjectMeta.Name)
					_, err := dst.Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), &crbl, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Hpa resource
		if stringInSlice("horizontalpodautoscaler", dst.Resources) || stringInSlice("horizontalpodautoscalers", dst.Resources) || stringInSlice("hpa", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating HorizontalPodAutoscalers")
			for _, hpa := range src_resources.HpaList {
				hpa := hpa
				if hpa.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating HorizontalPodAutoscaler: ", hpa.ObjectMeta.Name)
					_, err := dst.Clientset.AutoscalingV1().HorizontalPodAutoscalers(element.ObjectMeta.Name).Create(context.TODO(), &hpa, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Pod Security Policy resource
		if stringInSlice("podsecuritypolicies", dst.Resources) || stringInSlice("podsecuritypolicy", dst.Resources) || stringInSlice("psp", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Pod security policies")
			for _, psp := range src_resources.PspList {
				psp := psp
				if psp.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Pod Security Policies: ", psp.ObjectMeta.Name)
					_, err := dst.Clientset.PolicyV1beta1().PodSecurityPolicies().Create(context.TODO(), &psp, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}

		// Create Service Account resource
		if stringInSlice("serviceaccount", dst.Resources) || stringInSlice("serviceaccounts", dst.Resources) || stringInSlice("sa", dst.Resources) || stringInSlice("all", dst.Resources) {
			fmt.Println("===============")
			fmt.Println("Creating Service Account Job's")
			for _, sa := range src_resources.SvcAccList {
				sa := sa
				if sa.ObjectMeta.Namespace == element.ObjectMeta.Name {
					fmt.Println("Creating Service Account: ", sa.ObjectMeta.Name)
					_, err := dst.Clientset.CoreV1().ServiceAccounts(element.ObjectMeta.Name).Create(context.TODO(), &sa, metav1.CreateOptions{})
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
}

func Delete_resource_eks(dst *cluster.Cluster, src_resources *resource.Resources) {

	// Loop through each namespace and create resources inside each namespace
	for _, element := range src_resources.Nsl.Items {
        element := element
		fmt.Println("=====================================================================")
		fmt.Println("Operating on namespace: ", element.ObjectMeta.Name)
		fmt.Println("=====================================================================")
		// Delete PVC resource
		for _, pvc := range src_resources.PersistentVolumeClaimsList {
			if pvc.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting PVC: ", pvc.ObjectMeta.Name)
				err := dst.Clientset.CoreV1().PersistentVolumeClaims(element.ObjectMeta.Name).Delete(context.TODO(), pvc.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete Deployment resource
		fmt.Println("Deleting Deployments")
		for _, dep := range src_resources.Depl {
			dep := dep
			if dep.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Secret: ", dep.ObjectMeta.Name)
				err := dst.Clientset.AppsV1().Deployments(element.ObjectMeta.Name).Delete(context.TODO(), dep.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete Service resource
		fmt.Println("Deleting Services")
		for _, svc := range src_resources.Svcl {
			svc := svc
			if svc.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Service: ", svc.ObjectMeta.Name)
				err := dst.Clientset.CoreV1().Services(element.ObjectMeta.Name).Delete(context.TODO(), svc.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete DaemonSets resource
		fmt.Println("Deleting DaemonSets")
		for _, ds := range src_resources.Dsl {
		    ds := ds
			if ds.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting DaemonSet: ", ds.ObjectMeta.Name)
				err := dst.Clientset.AppsV1().DaemonSets(element.ObjectMeta.Name).Delete(context.TODO(), ds.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete secrets resource
		fmt.Println("===============")
		fmt.Println("Deleting Secrets")
		for _, secret := range src_resources.SecretList {
			secret := secret
			if secret.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Secret: ", secret.ObjectMeta.Name)
				err := dst.Clientset.CoreV1().Secrets(element.ObjectMeta.Name).Delete(context.TODO(), secret.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete StorageClass resource
		fmt.Println("===============")
		fmt.Println("Deleting StorageClass")
		for _, sc := range src_resources.StorageClassList {
			sc := sc
			if sc.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting StorageClass: ", sc.ObjectMeta.Name)
				err := dst.Clientset.StorageV1().StorageClasses().Delete(context.TODO(), sc.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		//Create MutatingWebhookConfiguration resource
		fmt.Println("===============")
		fmt.Println("Deleting MutatingWebhookConfiguration")
		for _, mwc := range src_resources.MutatingWebhookConfigurationList {
			mwc := mwc
			if mwc.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting MutatingWebhookConfiguration: ", mwc.ObjectMeta.Name)
				err := dst.Clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.TODO(), mwc.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete Configmap resource
		fmt.Println("===============")
		fmt.Println("Deleting ConfigMap's")
		for _, cm := range src_resources.ConfigMapsList {
			cm := cm
			if cm.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting ConfigMap: ", cm.ObjectMeta.Name)
				err := dst.Clientset.CoreV1().ConfigMaps(element.ObjectMeta.Name).Delete(context.TODO(), cm.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete CronJob resource
		fmt.Println("===============")
		fmt.Println("Deleting CronJob's")
		for _, cronjob := range src_resources.CronJobList {
			cronjob := cronjob
			if cronjob.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting CronJob: ", cronjob.ObjectMeta.Name)
				err := dst.Clientset.BatchV1beta1().CronJobs(element.ObjectMeta.Name).Delete(context.TODO(), cronjob.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete Job resource
		fmt.Println("===============")
		fmt.Println("Deleting Job's")
		for _, job := range src_resources.JobList {
			job := job
			if job.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Job: ", job.ObjectMeta.Name)
				err := dst.Clientset.BatchV1().Jobs(element.ObjectMeta.Name).Delete(context.TODO(), job.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete ingress resource
		fmt.Println("===============")
		fmt.Println("Deleting Ingresses")
		for _, ingress := range src_resources.IngressList {
			ingress := ingress
			if ingress.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Ingress: ", ingress.ObjectMeta.Name)
				err := dst.Clientset.NetworkingV1().Ingresses(element.ObjectMeta.Name).Delete(context.TODO(), ingress.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete HPA resource
		fmt.Println("===============")
		fmt.Println("Deleting HorizontalPodAutoscalers")
		for _, hpa := range src_resources.HpaList {
			hpa := hpa
			if hpa.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting HorizontalPodAutoscaler: ", hpa.ObjectMeta.Name)
				err := dst.Clientset.AutoscalingV1().HorizontalPodAutoscalers(element.ObjectMeta.Name).Delete(context.TODO(), hpa.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete psp resources
		fmt.Println("===============")
		fmt.Println("Deleting Pod Security Policies")
		for _, psp := range src_resources.PspList {
			psp := psp
			if psp.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting PodSecurityPolicy: ", psp.ObjectMeta.Name)
				err := dst.Clientset.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), psp.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		// Delete roles resource
		fmt.Println("===============")
		fmt.Println("Deleting Roles")
		for _, role := range src_resources.RoleList {
			role := role
			if role.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Roles: ", role.ObjectMeta.Name)
				err := dst.Clientset.RbacV1().Roles(element.ObjectMeta.Name).Delete(context.TODO(), role.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}
		// Delete service account resources
		fmt.Println("===============")
		fmt.Println("Deleting Service Accounts")
		for _, sa := range src_resources.SvcAccList {
			sa := sa
			if sa.ObjectMeta.Namespace == element.ObjectMeta.Name {
				fmt.Println("Deleting Service Accounts: ", sa.ObjectMeta.Name)
				err := dst.Clientset.CoreV1().ServiceAccounts(element.ObjectMeta.Name).Delete(context.TODO(), sa.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}

	}
	// Delete list of namespaces in destination cluster
	for _, element := range src_resources.Nsl.Items {

		element := element
		err := dst.Clientset.CoreV1().Namespaces().Delete(context.TODO(), element.ObjectMeta.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Println(err)
		}
	}
}

func Deploy_helm_charts(dst *cluster.Cluster, src_resources *resource.Resources) {
	args := map[string]string{}
	//repo 		:= ""

	for namespace, charts := range src_resources.HelmList {

		for key, value := range charts {
			//name := key
			//chart := value
			fmt.Println("Installing Chart ", key, " on EKS cluster in namespace ", namespace)
			//InstallChart(key,"", value, args, namespace)
			namespace = namespace
			args = args
			// Resolve chart dependency
			cmd := exec.Command("helm", "dependency", "build")
			cmd.Dir = value
			out, err := cmd.Output()
			if err != nil {
				fmt.Println("test2")
				log.Fatal(err)
			}
			
			fmt.Printf(" %s\n", out)
			//install charts
			cmd = exec.Command("helm", "upgrade", "--install", key, "." , "-n", namespace)
			cmd.Dir = value
			out, err = cmd.Output()
			if err != nil {
				fmt.Println("Error installing Helm chart. If there is a helm chart already on target cluster with name ", key, " in failed state try deleting and run again")
				log.Fatal(err)
			}
			fmt.Printf(" %s\n", out)

		}
	}
}

func Delete_helm_charts(dst *cluster.Cluster, src_resources *resource.Resources) {
	args := map[string]string{}
	//repo 		:= ""

	for namespace, charts := range src_resources.HelmList {

		for key, value := range charts {
			//name := key
			//chart := value
			fmt.Println("Uninstalling Chart ", key, " on EKS cluster")
			//InstallChart(key,"", value, args, namespace)
			namespace = namespace
			args = args
			//value = value

			//install charts
			cmd := exec.Command("helm", "uninstall", key, "-n", namespace)
			cmd.Dir = value
			out, err := cmd.Output()
			if err != nil {
				fmt.Println("Failed uninstalling chart ", key, " but continuing")
			}
			fmt.Printf(" %s\n", out)

		}
	}
}

// InstallChart
func InstallChart(name, repo, chart string, args map[string]string, namespace string) {
	//args        := map[string]string{}
	//repo 		:= ""

	settings = cli.New()

	//ns := namespace

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), debug); err != nil {
		log.Fatal(err)
	}
	client := action.NewInstall(actionConfig)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	//name, chart, err := client.NameAndChart(args)
	client.ReleaseName = name
	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repo, chart), settings)
	if err != nil {
		log.Fatal(err)
	}

	debug("CHART PATH: %s\n", cp)

	p := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		log.Fatal(err)
	}

	// Add args
	if err := strvals.ParseInto(args["set"], vals); err != nil {
		log.Fatal(errors.Wrap(err, "failed parsing --set data"))
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		log.Fatal(err)
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		log.Fatal(err)
	}

	// IF charts have dependencies try to install it
	for _, dep := range chartRequested.Metadata.Dependencies {
		if dep.Enabled == true {
			fmt.Println("Chart Dependencies :", dep.Name)
			RepoAdd(dep.Name, dep.Repository)
			RepoUpdate()
			//InstallChart(dep.Name, dep.Repository,"",args,namespace)
		}
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		fmt.Println("Checking Deps")
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			fmt.Println("Downloading Deps")
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
	}

	//fmt.Println("===============2============")

	client.Namespace = namespace
	release, err := client.Run(chartRequested, vals)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(release.Manifest)
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf("[debug] %s\n", format)
	log.Output(2, fmt.Sprintf(format, v...))
}

// RepoAdd adds repo with given name and url
func RepoAdd(name, url string) {
	repoFile := settings.RepositoryConfig

	//Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		log.Fatal(err)
	}

	if f.Has(name) {
		fmt.Printf("repository name (%s) already exists\n", name)
		return
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		log.Fatal(err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		log.Fatal(err)
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q has been added to your repositories\n", name)
}

// RepoUpdate updates charts for all helm repos
func RepoUpdate() {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		log.Fatal(errors.New("no repositories found. You must add one before updating"))
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			log.Fatal(err)
		}
		repos = append(repos, r)
	}

	fmt.Printf("Hang tight while we grab the latest from your chart repositories...\n")
	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				fmt.Printf("...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				fmt.Printf("...Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	fmt.Printf("Update Complete. ⎈ Happy Helming!⎈\n")
}
