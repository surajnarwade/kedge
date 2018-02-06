/*
Copyright 2017 The Kedge Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/asottile/dockerfile"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var (
	fileName, image, name string
	controller            string
	ports                 []string
)

const DockerfileName = "Dockerfile"

/*
**NOTE** to kedge devs:

The structs are re-defined here because, if we use the structs defined in `types.go`
it will clutter the output. Lot of upstream structs from OpenShift don't have
json tag `omitempty` defined on it's fields which causes lot of extra fields in yaml
output with zero values.

So redefining it here which helps us control how the output looks like. This can cause the
`types.go` and the structs defined here going out of sync if any major changes are done
to spec in types.go.
*/

type Container struct {
	Image string `json:"image,omitempty"`
}

type Service struct {
	PortMappings []string `json:"portMappings,omitempty"`
}

type App struct {
	Name       string      `json:"name,omitempty"`
	Controller string      `json:"controller,omitempty"`
	Containers []Container `json:"containers,omitempty"`
	Services   []Service   `json:"services,omitempty"`
}

// initCmd represents the version command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Kedge file",
	Run: func(cmd *cobra.Command, args []string) {

		// check if the file already exists
		_, err := os.Stat(fileName)
		if err == nil {
			fmt.Println(fileName, "is already present")
			os.Exit(-1)
		}

		// mandatory fields check
		if name == "" || image == "" {
			fmt.Println("--name and --image are mandatory flags, Please provide these flags")
			os.Exit(-1)
		}
		obj := App{
			Name:       name,
			Containers: []Container{{Image: image}},
		}

		if len(ports) > 0 {
			obj.Services = []Service{{PortMappings: ports}}
		} else {
			// Dockerfile detection
			_, err = os.Stat(DockerfileName)
			if err != nil {
				fmt.Println(DockerfileName, "is not present")
				os.Exit(-1)
			} else {
				a, err := dockerfile.ParseFile("Dockerfile")
				if err != nil {
					fmt.Println("Dockerfile parsing failed")
					os.Exit(-1)
				}
				for _, j := range a {
					if j.Cmd == "expose" {
						obj.Services = []Service{{PortMappings: j.Value}}
					}
				}
			}

		}

		// this switch is to check if user is not giving any wrong values
		switch strings.ToLower(controller) {
		case "deployment", "job", "deploymentconfig", "":
			obj.Controller = controller
		default:
			fmt.Println("'--controller' can only have values [Deployment, Job, DeploymentConfig].")
			os.Exit(-1)
		}

		// convert the internal form to yaml
		data, err := yaml.Marshal(obj)
		if err != nil {
			os.Exit(1)
		}

		f, err := os.Create(fileName)
		if err != nil {
			fmt.Println(err, "failed to create file")
			os.Exit(-1)
		}
		defer f.Close()

		// dump all the converted data into file
		_, err = f.Write(data)
		if err != nil {
			os.Exit(1)
		}
		fmt.Println("file", fileName, "created")

	},
}

func init() {
	initCmd.Flags().StringVarP(&fileName, "out", "o", "kedge.yml", "Output filename")
	initCmd.Flags().StringVarP(&name, "name", "n", "", "The name of service")
	initCmd.Flags().StringVarP(&image, "image", "i", "", "The image for the container to run")
	initCmd.Flags().StringSliceVarP(&ports, "ports", "p", []string{}, "The ports that this container exposes")
	initCmd.Flags().StringVarP(&controller, "controller", "c", "", "The type of controller this application is. Legal values [Deployment, Job, DeploymentConfig]. Default 'Deployment'.")
	RootCmd.AddCommand(initCmd)
}
