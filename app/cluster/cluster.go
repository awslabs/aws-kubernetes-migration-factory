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

 
package cluster

import (
	"os"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// establish connection with ks8
type Cluster struct {
	Kubeconfig_path string                // Path to the kubeconfig file
	Clientset       *kubernetes.Clientset // Client pointing the CKE cluster
	Region          string                // GCP region in which the cluster is running
	Namespaces      []string              // namespaces in kubernetes cluster from which the resources will be scanned
	Context         string                // context of Kubeconfig file
	Resources       []string              // Resources to include
	Helm_path       string                // Path to save helm path on local system
	Migrate_Images  string                // Migrate images from 3rd party registries to ECR
    Registry_Names  []string              // List of 3rd party registry names

}

func (c *Cluster) SetKubeconfig_path(kubeconfig_path string) {
    c.Kubeconfig_path = kubeconfig_path
}

func (c Cluster) GetKubeconfig_path() string {
    return c.Kubeconfig_path
}

func (c *Cluster) SetClientset(clientset *kubernetes.Clientset) {
    c.Clientset = clientset
}

func (c Cluster) GetClientset() *kubernetes.Clientset {
    return c.Clientset
}

func (c *Cluster) SetRegion(region string) {
    c.Region = region
}

func (c Cluster) GetRegion() string {
    return c.Region
}

func (c *Cluster) SetNamespaces(namespaces []string) {
    c.Namespaces = namespaces
}

func (c Cluster) GetNamespaces() []string {
    return c.Namespaces
}

func (c *Cluster) SetContext(context string) {
    c.Context = context
}

func (c Cluster) GetContext() string {
    return c.Context
}

func (c *Cluster) SetResources(resources []string) {
    c.Resources = resources
}

func (c Cluster) GetResources() []string {
    return c.Resources
}

func (c *Cluster) SetHelm_path(helm_path string) {
    c.Helm_path = helm_path
}

func (c Cluster) GetHelm_path() string {
    return c.Helm_path
}

func (c *Cluster) SetMigrate_Images(migrate_images string) {
    c.Migrate_Images = migrate_images
}

func (c Cluster) GetMigrate_Images() string {
    return c.Migrate_Images
}

func (c *Cluster) SetRegistry_Names(reg_names string) {
    c.Registry_Names = append(c.Registry_Names, reg_names)
}

func (c Cluster) GetRegistry_Names() []string {
    return c.Registry_Names
}

// retrieve cluster client
func get_cluster_client(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

// Generate client for the source cluster config passed
func (c *Cluster) Generate_cluster_client() {
	//c.Clientset = clientset
	
	config, err := get_cluster_client(c.Context, c.Kubeconfig_path)
	if err != nil {
		fmt.Printf("The kubeconfig cannot be loaded: %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	c.SetClientset ( clientset )
}
