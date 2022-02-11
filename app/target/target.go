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


package source

import (
	cluster "containers-migration-factory/app/cluster"
	resource "containers-migration-factory/app/resource"
)

// Geometry is an interface that defines Geometrical Calculation
type Target interface {
	Connect(sCluster *cluster.Cluster)
	// GetSourceDetails(sCluster *cluster.Cluster) resource.Resources 
	DeployResources(sCluster *cluster.Cluster,srcResources *resource.Resources,action string)
	FormatSourceData (resource *resource.Resources) // trim / data clean up
}

func SetContext(target Target,dCluster *cluster.Cluster){
	/*Connect to target cluster*/
	target.Connect(dCluster)
}

// Invoke specific source based on input provided
func Invoke(target Target, sType string, sCluster *cluster.Cluster, dCluster *cluster.Cluster,srcResources *resource.Resources, action string) string {
	
	/*Connect to target cluster*/
	// target.Connect(dCluster)

	target.DeployResources(dCluster,srcResources,action)

	/*Get Source Details*/

	// resources := source.GetSourceDetails(sCluster)

	// source.FormatSourceData(&resources)

	return "success"
}
