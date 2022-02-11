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

package MIGRATE_IMAGES

import (
	"strings"
	"fmt"
	"regexp"
	"os/exec"
	"log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/sts"
)

type GetCallerIdentityOutput struct {
    // The AWS account ID number of the account that owns or contains the calling
    // entity.
    Account *string `type:"string"`

    // The AWS ARN associated with the calling entity.
    Arn *string `min:"20" type:"string"`

    // The unique identifier of the calling entity. The exact value depends on the
    // type of entity that is making the call. The values returned are those listed
    // in the aws:userid column in the Principal table (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_variables.html#principaltable)
    // found on the Policy Variables reference page in the IAM User Guide.
    UserId *string `type:"string"`
}

func create_ecr_repo(src_repo_name string, aws_region string) (ecr_create_status string){
	svc := ecr.New(session.New(&aws.Config{Region: aws.String(aws_region)}))
	input := &ecr.CreateRepositoryInput{
	    RepositoryName: aws.String(src_repo_name),
	}

	result, err := svc.CreateRepository(input)
	if err != nil {
	    if aerr, ok := err.(awserr.Error); ok {
	        switch aerr.Code() {
	        case ecr.ErrCodeServerException:
	            fmt.Println(ecr.ErrCodeServerException, aerr.Error())
	        case ecr.ErrCodeInvalidParameterException:
	            fmt.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
	        case ecr.ErrCodeInvalidTagParameterException:
	            fmt.Println(ecr.ErrCodeInvalidTagParameterException, aerr.Error())
	        case ecr.ErrCodeTooManyTagsException:
	            fmt.Println(ecr.ErrCodeTooManyTagsException, aerr.Error())
	        case ecr.ErrCodeRepositoryAlreadyExistsException:
	            fmt.Println(ecr.ErrCodeRepositoryAlreadyExistsException, aerr.Error())
	        case ecr.ErrCodeLimitExceededException:
	            fmt.Println(ecr.ErrCodeLimitExceededException, aerr.Error())
	        default:
	            fmt.Println(aerr.Error())
	        }
	    } else {
	        // Print the error, cast err to awserr.Error to get the Code and
	        // Message from an error.
	        fmt.Println(err.Error())
	    }
	    return
	}

	return *result.Repository.RepositoryUri
}

func list_ecr_repo(aws_region string) (ecr_repo_list []string) {
        svc := ecr.New(session.New(&aws.Config{Region: aws.String(aws_region)}))
        input := &ecr.DescribeRepositoriesInput{}

        result, err := svc.DescribeRepositories(input)
        if err != nil {
                if aerr, ok := err.(awserr.Error); ok {
                        switch aerr.Code() {
                        case ecr.ErrCodeServerException:
                                fmt.Println(ecr.ErrCodeServerException, aerr.Error())
                        case ecr.ErrCodeInvalidParameterException:
                                fmt.Println(ecr.ErrCodeInvalidParameterException, aerr.Error())
                        case ecr.ErrCodeRepositoryNotFoundException:
                                fmt.Println(ecr.ErrCodeRepositoryNotFoundException, aerr.Error())
                        default:
                                fmt.Println(aerr.Error())
                        }
                } else {
                        // Print the error, cast err to awserr.Error to get the Code and
                        // Message from an error.
                        fmt.Println(err.Error())
                }
		//return err
        }
	
	for _, repo_list := range result.Repositories {
		ecr_repo_list = append (ecr_repo_list, *repo_list.RepositoryUri)
	}
	return ecr_repo_list	
}

func STS_GetCallerIdentity() (aws_account string) {
	svc := sts.New(session.New())
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return 
	}
	aws_account = *result.Account
	return aws_account
}

func check_ecr_repo(src_image_name string, src_repo_name string, src_image_tag string) (updated_image_name string) {
	var validate_ecr string = ""
	var aws_region string
	var aws_account string
	//var stderr bytes.Buffer
	//var cmdout bytes.Buffer

	cmd := exec.Command("aws", "configure", "get", "region")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	aws_region = strings.TrimSuffix(string(out), "\n")
	aws_account = STS_GetCallerIdentity()
		
	first_join := []string{fmt.Sprint(aws_account), "dkr", "ecr", string(aws_region), "amazonaws", "com"}
	ecr_reg_stage := strings.Join(first_join, ".")
	second_join := []string{ecr_reg_stage, src_repo_name}
	ecr_reg_url := strings.Join(second_join, "/")
	collect_ecr_reg_url := list_ecr_repo(aws_region)
	
	for _, list := range collect_ecr_reg_url {
		if list == ecr_reg_url {
			validate_ecr = "true"
		}
	}

	if validate_ecr != "true" {
		ecr_repo_name := create_ecr_repo(src_repo_name, aws_region)
		fmt.Println("Successfully created ECR repository named :", ecr_repo_name)
	}
		
	updated_image_stage := []string{ecr_reg_url, src_image_tag}
	updated_image_name = strings.Join(updated_image_stage, ":")	

	srcimage_pull_cmd := exec.Command("docker", "pull", src_image_name)	
	srcimagepullout, err := srcimage_pull_cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(srcimagepullout))
		log.Fatal(err)
	}
	fmt.Printf("%s \n", srcimagepullout)

	image_tag_cmd := exec.Command("docker", "tag", src_image_name, updated_image_name)
	imagetagout, err := image_tag_cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(imagetagout))
		log.Fatal(err)
	}
	fmt.Printf("%s \n", imagetagout)

	dstimage_push_cmd := exec.Command("docker", "push", updated_image_name)
	imagepushout, err := dstimage_push_cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(imagepushout))
		log.Fatal(err)
	}
	fmt.Printf("%s \n", imagepushout)

	return updated_image_name
		
}

func Validate(src_image_name string, external_reg_names []string) (updated_image string) {
	updated_image = ""
	var src_image_tag string
	var src_registry_url string
        var src_registry_host string

	check_src_image_reg := strings.Split(src_image_name, ":")
	if len(check_src_image_reg) == 1 {
		src_image_tag = "latest"
		src_registry_url = check_src_image_reg[len(check_src_image_reg)-1]
	} else if len(check_src_image_reg) == 2 {
		src_image_tag = check_src_image_reg[len(check_src_image_reg)-1]
		src_registry_url = check_src_image_reg[len(check_src_image_reg)-2]
	}
	
	src_registry_url_split := strings.Split(src_registry_url, "/")

	if len(src_registry_url_split) != 1 {
		src_registry_host = src_registry_url_split[len(src_registry_url_split)-len(src_registry_url_split)]
		re := regexp.MustCompile(`\.`)
		regex_result := re.Match([]byte(src_registry_host))
		if regex_result != true {
			src_registry_host = "dockerhub"
		}
	}

	for _, url := range external_reg_names {
		if src_registry_host == url {
			src_repo_name := strings.Join(src_registry_url_split[1:], "/")
			updated_image = check_ecr_repo(src_image_name, src_repo_name, src_image_tag)
		} 
	}
	return updated_image
}
