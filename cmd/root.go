/*
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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/eks-node-viewer/pkg/pricing"
	"github.com/imdario/mergo"
	"github.com/olekukonko/tablewriter"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	OutputYAML       = "yaml"
	OutputTableShort = "short"
	OutputTableWide  = "wide"
)

var (
	version = ""
)

type GlobalOptions struct {
	Verbose           bool
	Version           bool
	Output            string
	ConfigFile        string
	Region            string
	Replacement       string
	Flexibility       string
	CapacityType      string
	PricingMultiplier float64
}

var (
	globalOpts = GlobalOptions{}
	rootCmd    = &cobra.Command{
		Use:     "ec2-replacement-sim",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
			if globalOpts.Region != "" {
				sess.Config.Region = &globalOpts.Region
			}
			flexibilityRegex := regexp.MustCompile(globalOpts.Flexibility)
			pricing.NewPricingAPI(sess, *sess.Config.Region)
			pricesUpdated := make(chan struct{})
			pricingProvider := pricing.NewProvider(cmd.Context(), sess, func() {
				pricesUpdated <- struct{}{}
			})
			fmt.Println("Waiting for pricing data to be pulled")
			<-pricesUpdated
			instanceTypeToReplacePrice, ok := pricingProvider.SpotPrice(globalOpts.Replacement, firstZone(*sess.Config.Region))
			if !ok {
				panic(fmt.Sprintf("can't find pricing info for %s", globalOpts.Replacement))
			}
			var flexibilitySet []string
			for _, it := range pricingProvider.InstanceTypes() {
				if flexibilityRegex.Match([]byte(it)) {
					flexibilitySet = append(flexibilitySet, it)
				}
			}
			sort.Slice(flexibilitySet, func(i, j int) bool {
				iPrice, _ := pricingProvider.SpotPrice(flexibilitySet[i], firstZone(*sess.Config.Region))
				jPrice, _ := pricingProvider.SpotPrice(flexibilitySet[j], firstZone(*sess.Config.Region))
				return iPrice < jPrice
			})
			thresholdPrice := instanceTypeToReplacePrice * globalOpts.PricingMultiplier
			var replacementCandidates []string
			for _, it := range flexibilitySet {
				price, ok := pricingProvider.SpotPrice(it, firstZone(*sess.Config.Region))
				if !ok {
					log.Printf("Not able to find pricing for %s, skipping", it)
					continue
				}
				if price < thresholdPrice {
					replacementCandidates = append(replacementCandidates, it)
				} else if globalOpts.Verbose {
					log.Printf("%s ($%.3f) was not below the pricing threshold of $%.3f:", it, price, thresholdPrice)
				}
			}
			fmt.Printf("Replacement Instance Type: %s\n", globalOpts.Replacement)
			fmt.Printf("      Instance Type Price: $%.3f\n", instanceTypeToReplacePrice)
			fmt.Printf("          Threshold Price: $%.3f\n", thresholdPrice)
			fmt.Printf("          Flexibility Set: %d\n", len(flexibilitySet))
			fmt.Printf("   Replacement Candidates: %d\n", len(replacementCandidates))
			for _, it := range replacementCandidates {
				fmt.Printf("  - %s\n", it)
			}
		},
	}
)

func firstZone(region string) string {
	return fmt.Sprintf("%sa", region)
}

func main() {
	rootCmd.PersistentFlags().Float64Var(&globalOpts.PricingMultiplier, "pricing-multiplier", 0.5, "Pricing Multipler to determine replacement threshold")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Flexibility, "flexibility", `^(c|m|r).[a-z0-9]+$`, "Flexibility Set (regex)")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Replacement, "replacement", "", "Replacement Instance Type")
	rootCmd.PersistentFlags().StringVar(&globalOpts.CapacityType, "capacity-type", "spot", "Capacity Type (spot or on-demand)")
	rootCmd.PersistentFlags().StringVar(&globalOpts.Region, "region", "", "AWS Region")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.Verbose, "verbose", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&globalOpts.Version, "version", false, "version")
	rootCmd.PersistentFlags().StringVarP(&globalOpts.Output, "output", "o", OutputTableShort,
		fmt.Sprintf("Output mode: %v", []string{OutputTableShort, OutputTableWide, OutputYAML}))
	rootCmd.PersistentFlags().StringVarP(&globalOpts.ConfigFile, "file", "f", "", "YAML Config File")

	rootCmd.AddCommand(&cobra.Command{Use: "completion", Hidden: true})
	cobra.EnableCommandSorting = false

	lo.Must0(rootCmd.Execute())
}

func ParseConfig[T any](globalOpts GlobalOptions, opts T) (T, error) {
	if globalOpts.ConfigFile == "" {
		return opts, nil
	}
	configBytes, err := os.ReadFile(globalOpts.ConfigFile)
	if err != nil {
		return opts, err
	}
	var parsedCreateOpts T
	if err := yaml.Unmarshal(configBytes, &parsedCreateOpts); err != nil {
		return opts, err
	}
	if err := mergo.Merge(&opts, parsedCreateOpts, mergo.WithOverride); err != nil {
		return opts, err
	}
	return opts, nil
}

func PrettyEncode(data any) string {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		panic(err)
	}
	return buffer.String()
}

func PrettyTable[T any](data []T, wide bool) string {
	var headers []string
	var rows [][]string
	for _, dataRow := range data {
		var row []string
		// clear headers each time so we only keep one set
		headers = []string{}
		reflectStruct := reflect.Indirect(reflect.ValueOf(dataRow))
		for i := 0; i < reflectStruct.NumField(); i++ {
			typeField := reflectStruct.Type().Field(i)
			tag := typeField.Tag.Get("table")
			if tag == "" {
				continue
			}
			subtags := strings.Split(tag, ",")
			if len(subtags) > 1 && subtags[1] == "wide" && !wide {
				continue
			}
			headers = append(headers, subtags[0])
			row = append(row, reflect.ValueOf(dataRow).Field(i).String())
		}
		rows = append(rows, row)
	}
	out := bytes.Buffer{}
	table := tablewriter.NewWriter(&out)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows) // Add Bulk Data
	table.Render()
	return out.String()
}
