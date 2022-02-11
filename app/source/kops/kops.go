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


package kops

import (
	cluster "containers-migration-factory/app/cluster"
	resource "containers-migration-factory/app/resource"
	"fmt"
)

// kops defines as source
type KOPS struct{}

func (c KOPS) Connect(sCluster *cluster.Cluster) {
	sCluster.Generate_cluster_client()
}

func (k KOPS) GetSourceDetails(sCluster *cluster.Cluster) resource.Resources {
	fmt.Println("KOPS GetSourceDetails....")
	resources := resource.Resources{}

	return resources
}

// GCP FormatSourceData implements the Geometry interface
func (k KOPS) FormatSourceData(resource *resource.Resources, resToInclude []string) {
}
