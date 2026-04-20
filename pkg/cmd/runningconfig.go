/*
Copyright 2018 The Kubernetes Authors.

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
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sdcio/kubectl-sdc/pkg/client"
	"github.com/sdcio/kubectl-sdc/pkg/commands/runningconfig"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

type RunningConfigOptions struct {
	target    string
	formatStr string
	format    client.Format
	GenericOptions
}

// NewRunningConfigOptions provides an instance of RunningConfigOptions with default values
func NewRunningConfigOptions(streams genericiooptions.IOStreams) *RunningConfigOptions {
	return &RunningConfigOptions{
		GenericOptions: GenericOptions{
			configFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

func (o *RunningConfigOptions) Complete(_ *cobra.Command, _ []string) error {
	var err error
	clientConfig := o.configFlags.ToRawKubeConfigLoader()

	o.restConfig, err = o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	// retrieve the actual namespace from clientConfig
	o.namespace, _, err = clientConfig.Namespace()
	if err != nil {
		return err
	}

	return nil
}

// Validate validates the options
func (o *RunningConfigOptions) Validate() error {
	if o.target == "" {
		return fmt.Errorf("target not set")
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace not set")
	}
	// Parse format string
	format, err := runningconfig.ParseFormat(o.formatStr)
	if err != nil {
		return err
	}
	o.format = format
	return nil
}

func (o *RunningConfigOptions) Run(_ *cobra.Command) error {
	ctx := context.Background()

	cl, err := client.NewConfigClient(o.restConfig)
	if err != nil {
		return err
	}
	output, err := runningconfig.Run(ctx, cl, o.namespace, o.target, o.format)
	if err != nil {
		return err
	}

	// Display the formatted output
	_, err = fmt.Fprintln(o.Out, output)
	return err
}

// NewCmdRunningConfig provides a cobra command wrapping RunningConfigOptions
func NewCmdRunningConfig(streams genericiooptions.IOStreams) (*cobra.Command, error) {

	o := NewRunningConfigOptions(streams)
	cmd := &cobra.Command{
		Use:          "runningconfig",
		Short:        "Get running configuration for a target",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {

			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(c); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&o.target, "target", "", "target to get the running config for")
	err := cmd.MarkFlagRequired("target")
	if err != nil {
		return nil, err
	}

	// Build format help text dynamically
	formatHelp := fmt.Sprintf("output format (%s)", runningconfig.FormatListString())
	cmd.Flags().StringVar(&o.formatStr, "format", "yaml", formatHelp)

	// Format flag completion
	if err := cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return runningconfig.ValidFormatStrings(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		return nil, err
	}

	if err := cmd.RegisterFlagCompletionFunc("target", targetCompletionFunc(o)); err != nil {
		return nil, err
	}
	o.configFlags.AddFlags(cmd.Flags())

	return cmd, nil
}
