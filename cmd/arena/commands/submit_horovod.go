// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	horovod_training_chart = util.GetChartsFolder() + "/tf-horovod"
)

// NewSubmitHorovodJobCommand
func NewSubmitHorovodJobCommand() *cobra.Command {
	var (
		submitArgs submitHorovodJobArgs
	)

	submitArgs.Mode = "horovodjob"

	var command = &cobra.Command{
		Use:     "horovodjob",
		Short:   "Submit horovodjob as training job.",
		Aliases: []string{"hj"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitHorovodJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().StringVar(&submitArgs.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&submitArgs.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")

	submitArgs.addCommonFlags(command)
	submitArgs.addSyncFlags(command)

	command.Flags().IntVar(&submitArgs.SSHPort, "sshPort", 0,
		"ssh port.")
	return command
}

type submitHorovodJobArgs struct {
	SSHPort int    `yaml:"sshPort"` // --sshPort
	Cpu     string `yaml:"cpu"`     // --cpu
	Memory  string `yaml:"memory"`  // --memory

	// for common args
	submitArgs `yaml:",inline"`

	// for tensorboard
	submitTensorboardArgs `yaml:",inline"`

	// for sync up source code
	submitSyncCodeArgs `yaml:",inline"`
}

func (submitArgs *submitHorovodJobArgs) prepare(args []string) (err error) {
	submitArgs.Command = strings.Join(args, " ")

	sshPort, err := util.SelectAvailablePortWithDefault(clientset, submitArgs.SSHPort)
	if err != nil {
		return fmt.Errorf("failed to select ssh port: %++v", err)
	}
	submitArgs.SSHPort = sshPort

	err = submitArgs.check()
	if err != nil {
		return err
	}

	commonArgs := &submitArgs.submitArgs
	err = commonArgs.transform()
	if err != nil {
		return nil
	}

	if len(envs) > 0 {
		submitArgs.Envs = transformSliceToMap(envs, "=")
	}

	submitArgs.addHorovodInfoToEnv()

	return nil
}

func (submitArgs submitHorovodJobArgs) check() error {
	err := submitArgs.submitArgs.check()
	if err != nil {
		return err
	}

	if submitArgs.Image == "" {
		return fmt.Errorf("--image must be set ")
	}

	return nil
}

func (submitArgs *submitHorovodJobArgs) addHorovodInfoToEnv() {
	submitArgs.addJobInfoToEnv()
}

func submitHorovodJob(args []string, submitArgs *submitHorovodJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	trainer := NewHorovodJobTrainer(clientset)
	job, err := trainer.GetTrainingJob(name, namespace)
	if err != nil {
		log.Debugf("Check %s exist due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s already exists, please delete it first via 'arena delete %s'", name, name)
	}

	err = workflow.SubmitJob(name, submitArgs.Mode, namespace, submitArgs, horovod_training_chart)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can check the job status via `arena get %s --type %s`", name, submitArgs.Mode)
	return nil
}

func submitHorovodJobWithHelm(args []string, submitArgs *submitHorovodJobArgs) (err error) {
	err = submitArgs.prepare(args)
	if err != nil {
		return err
	}

	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s already exists, please delete it first via 'arena delete %s'", name, name)
	}

	// the master is also considered as a worker
	submitArgs.WorkerCount = submitArgs.WorkerCount - 1

	return helm.InstallRelease(name, namespace, submitArgs, horovod_training_chart)
}
