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

 
package main

import (
	"bufio"
	//AWS "containers-migration-factory/controllers/AWS"
	//GCP "containers-migration-factory/controllers/GCP"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"os"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
	"github.com/bigkevmcd/go-configparser"

	gke "containers-migration-factory/app/source/gke"
	aks "containers-migration-factory/app/source/aks"
	kops "containers-migration-factory/app/source/kops"
	cluster "containers-migration-factory/app/cluster"
	source "containers-migration-factory/app/source"
	resource "containers-migration-factory/app/resource"
	eks "containers-migration-factory/app/target/eks"
	target "containers-migration-factory/app/target"
)

type Config struct {
	CurrentContext string `yaml:"current-context"`
}

var config Config

func fileExists(path string) bool {
    _, err := os.Stat(path)
    return !os.IsNotExist(err)
}

func Get_user_input(reader *bufio.Reader) (cluster.Cluster, cluster.Cluster, string, string) {
	sourceCluster := cluster.Cluster{}
	destCluster := cluster.Cluster{}

	
	namespaces_param := ""
	resources_param := ""
	helm_path_param := ""
	action_param := ""
	source_kubeconfig_param := ""
	source_context_param := ""
	src_cloud := ""
	destination_kubeconfig_param := ""
	destination_context_param := ""
	migrate_images_param := ""
	reg_names_param := ""

	if fileExists("config.ini"){
		configParams, err := configparser.NewConfigParserFromFile("config.ini")
		if err != nil {
			fmt.Printf("Error opening the config file for parameters: %v\n", err)
		} else {
			// get common section
			common_options, err := configParams.Items("COMMON")
			if err == nil{
				namespaces_param = common_options["NAMESPACES"]
				resources_param = common_options["RESOURCES"]
				helm_path_param = common_options["HELM_CHARTS_PATH"]
				action_param = common_options["ACTION"]
			}
			
			// get source section
			source_options, err := configParams.Items("SOURCE")
			if err == nil{
				source_kubeconfig_param = source_options["KUBE_CONFIG"]
				source_context_param = source_options["CONTEXT"]
				src_cloud = source_options["CLOUD"]
			}

			// get target section
			target_options, err := configParams.Items("TARGET")
			if err == nil{
				destination_kubeconfig_param = target_options["KUBE_CONFIG"]
				destination_context_param = target_options["CONTEXT"]
				// target_cloud := target_options["CLOUD"]
			}

			// get image migration section
			migrate_image_options, err := configParams.Items("MIGRATE_IMAGES")
			if err == nil{
					migrate_images_param = migrate_image_options["USERCONSENT"]
					reg_names_param = migrate_image_options["REGISTRY"]
			}

		}
	} else{
		fmt.Printf("Config.ini file doesn't exist and will use the user arguments\n")
	}
	//// Accept Source Cluster input
	source_kubeconfig := flag.String("source_kubeconfig", source_kubeconfig_param, "a string")
	namespaces := flag.String("namespaces", namespaces_param, "a string")
	destination_kubeconfig := flag.String("destination_kubeconfig", destination_kubeconfig_param, "a string")
	source_context := flag.String("source_context", source_context_param, "a string")
	destination_context := flag.String("destination_context", destination_context_param, "a string")
	resources := flag.String("resources", resources_param, "a string")
	helm_path := flag.String("helm_path", helm_path_param, "Path on local system where Helm charts from source cluster will be stored")
	migrate_images := flag.String("migrate_images", migrate_images_param, "User consent for migrating image from 3rd party registries to ECR")
	reg_names := flag.String("reg_names", reg_names_param, "List of 3rd party registries as comma separated items")
	action := flag.String("action", action_param, "What action the tools needs to perform. Accepted values are Deploy or Delete")
	sourceType := flag.String("source_type", src_cloud, "What is source type. Accepted values are GKE,AKS,KOPS")
	flag.Parse()

	// SOURCE ===================
	if *source_kubeconfig == "" {
		fmt.Print("Please pass the location of source kubernetes cluster kubeconfig file: ")
		*source_kubeconfig, _ = reader.ReadString('\n')
		//source_kubeconfig = "/home/ec2-user/.kube/gke"
	}

	// get current source context
	current_src_context := get_current_context(strings.TrimSuffix(*source_kubeconfig, "\n"))

	if *source_context == "" {
		fmt.Printf("Please pass the source context (default: %v): ", current_src_context)
		*source_context, _ = reader.ReadString('\n')
	}
	sourceCluster.SetContext( strings.TrimSuffix(*source_context, "\n") )

	if *resources == "" {
		fmt.Printf("Please pass comma separated list of resources to migrate from source cluster to destination cluster. For all resources enter 'all': ")
		*resources, _ = reader.ReadString('\n')
	}

	if *namespaces == "" {
		fmt.Print("Please pass comma separated list of namespaces for source cluster. For all namespaces enter 'all': ")
		*namespaces, _ = reader.ReadString('\n')
	}

	if *helm_path == "" {
		fmt.Print("Please pass path to save Helm charts from source cluster: ")
		*helm_path, _ = reader.ReadString('\n')
	}

	sourceCluster.SetHelm_path ( strings.TrimSuffix(*helm_path, "\n") )
	destCluster.SetHelm_path ( strings.TrimSuffix(*helm_path, "\n") )

	// Remove the newline character from the end of filepath entered by user
	sourceCluster.SetKubeconfig_path ( strings.TrimSuffix(*source_kubeconfig, "\n") )
	sourceCluster.SetContext ( strings.TrimSuffix(*source_context, "\n") )

	// Migrate Images from source container registry to ECR
	if *migrate_images == "" {
		fmt.Print("Do you want to migrate images from 3rd party registries to ECR? Supply either Yes or No: ")
		*migrate_images, _ = reader.ReadString('\n')
	}

	// Remove the newline character from the end of filepath entered by user
	sourceCluster.SetMigrate_Images ( strings.TrimSuffix(*migrate_images, "\n"))

	if sourceCluster.GetMigrate_Images() == "Yes" || sourceCluster.GetMigrate_Images() == "yes" {
		if *reg_names == "" {
			fmt.Print("Tool supports migration from gcr, gitlab, dockerhub registries. Please pass comma separated list of 3rd party registries: ")
			*reg_names, _ = reader.ReadString('\n')
		}
	}

	regnames_list := strings.Split(stripSpaces(*reg_names), ",")
	for _, regitem := range regnames_list {
		if regitem == "GCR" || regitem == "gcr" {
			regurl := []string{"asia.gcr.io", "eu.gcr.io", "gcr.io", "marketplace.gcr.io", "staging-k8s.gcr.io", "us.gcr.io"}
			for _, gcreg := range regurl {
				sourceCluster.SetRegistry_Names ( strings.TrimSuffix(gcreg, "\n"))
			}
		} else if regitem == "DOCKERHUB" || regitem == "dockerhub" {
			regurl := "dockerhub"
			sourceCluster.SetRegistry_Names ( strings.TrimSuffix(regurl, "\n"))
		} else if regitem == "gitlab" || regitem == "GITLAB" {
			regurl := "registry.gitlab.com"
			sourceCluster.SetRegistry_Names ( strings.TrimSuffix(regurl, "\n"))
		}
	}

	*namespaces = strings.TrimSuffix(*namespaces, "\n")
	if *namespaces == "" {
		fmt.Println("Namespace value not passed, exiting")
		os.Exit(4)
	}
	if *namespaces != "all" {
		sourceCluster.SetNamespaces ( strings.Split(stripSpaces(*namespaces), ",") )
	}

	*resources = strings.TrimSuffix(*resources, "\n")
	if *resources != "" {
		sourceCluster.SetResources ( strings.Split(stripSpaces(*resources), ",") )
		destCluster.SetResources ( sourceCluster.GetResources() )
	}

	// DESTINATION ==============
	if *destination_kubeconfig == "" {
		fmt.Print("Please pass the location of destination EKS cluster kubeconfig file: ")
		*destination_kubeconfig, _ = reader.ReadString('\n')
	}

	// get current destination context
	current_dst_context := get_current_context(strings.TrimSuffix(*destination_kubeconfig, "\n"))

	if *destination_context == "" {
		fmt.Printf("Please pass the destination context (default: %v): ", current_dst_context)
		*destination_context, _ = reader.ReadString('\n')
	}

	// Action for the tool
	if *action != "Deploy" && *action != "Delete" {
		fmt.Print("Please pass what action the tool needs to perform. Accepted values are Deploy or Delete : ")
		*action, _ = reader.ReadString('\n')
		*action = strings.TrimSuffix(*action, "\n")
		//fmt.Println("action entered", *action)
		if *action != "Deploy" && *action != "Delete" {
			fmt.Print("Invalid input for parameter \"action\", accepted values are Deploy or Delete")
			os.Exit(1)
		}
	}

	// fmt.Println("Action", *action)

	
	
	if *sourceType == "" {
		fmt.Print("Please pass source type  (supported source types GKE,AKS,KOPS): ")
		*sourceType, _ = reader.ReadString('\n')
		*sourceType = strings.TrimRight(*sourceType, "\n")
	}

	destCluster.SetKubeconfig_path ( strings.TrimSuffix(*destination_kubeconfig, "\n") )
	destCluster.SetContext ( strings.TrimSuffix(*destination_context, "\n") )

	return sourceCluster, destCluster , *action, *sourceType
}

// Get default cluster in config
func get_current_context(kubeconfigPath string) (default_context string) {
	kubeconfigPath = filepath.Clean(kubeconfigPath)
	source, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		fmt.Printf("Error opening up kube config file: %v\n", err)
		os.Exit(1)
	}

	err = yaml.Unmarshal(source, &config)
	if err != nil {
		fmt.Printf("Error getting default context from source config: %v\n", err)
		os.Exit(1)
	}

	default_context = config.CurrentContext

	return default_context
}

// helper function to drop spaces
func stripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

func main() {

	reader := bufio.NewReader(os.Stdin)
	//accept user input
	sourceCluster, destCluster, action, sourceType := Get_user_input(reader)
	fmt.Println(action)

	/*Get connect with source and target clusters*/

	//fmt.Println("sourceType:::", *sourceType)
	g := new(gke.GKE)
	a := new(aks.AKS)
	k := new(kops.KOPS)
	t := new(eks.EKS)
	var sourceResources resource.Resources
	target.SetContext(t,&destCluster)

	if sourceType == "GKE"  {
		fmt.Println("GKE Resources")
		source.SetContext(g,&sourceCluster)
		sourceResources = source.Invoke( g , sourceType, &sourceCluster, &destCluster)
		// fmt.Println(sourceResources)
	} else if sourceType == "AKS" {
		source.SetContext(a,&sourceCluster)
		sourceResources = source.Invoke(a , sourceType, &sourceCluster, &destCluster )
		// fmt.Println(sourceResources)
	} else if sourceType == "KOPS" {
		source.SetContext(k,&sourceCluster)
		sourceResources = source.Invoke(k, sourceType, &sourceCluster, &destCluster )
		// fmt.Println(sourceResources)
	} else{
		fmt.Println("Invalid input for parameter \"sourceType\", accepted values are GKE,AKE,KOPS")
		os.Exit(1)
	}
	target.Invoke(t,sourceType, &sourceCluster, &destCluster,&sourceResources, action)
}
