package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

import (
	msconfig "mockserver/config"
)

const jsonSchemaUrl string = "https://opensource.trymagic.xyz/schemas/mockserver.schema.json"

type OrderedConfig struct {
	Schema string      `json:"$schema,omitempty" yaml:"$schema,omitempty"`
	Server interface{} `json:"server" yaml:"server"`
	Groups interface{} `json:"groups,omitempty" yaml:"groups,omitempty"`
	Routes interface{} `json:"routes" yaml:"routes"`
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert config file between YAML and JSON",
	Run: func(cmd *cobra.Command, args []string) {
		if inputFile == "" || outputFile == "" {
			fmt.Println("Both --input and --output are required")
			os.Exit(1)
		}

		cfgData, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("[ERROR] Failed to read input file: %v\n", err)
			os.Exit(1)
		}

		var cfg msconfig.Config
		ext := strings.ToLower(filepath.Ext(inputFile))
		cfg.Schema = jsonSchemaUrl
		switch ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(cfgData, &cfg); err != nil {
				fmt.Printf("[ERROR] Failed to parse YAML: %v\n", err)
				os.Exit(1)
			}
		case ".json":
			if err := json.Unmarshal(cfgData, &cfg); err != nil {
				fmt.Printf("[ERROR] Failed to parse JSON: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Println("[ERROR] Unsupported input file format. Use .yaml/.yml or .json")
			os.Exit(1)
		}

		var outData []byte
		outExt := strings.ToLower(filepath.Ext(outputFile))

		ordered := OrderedConfig{
			Schema: cfg.Schema,
			Server: removeEmptyFields(cfg.Server),
			Routes: removeEmptyFields(cfg.Routes),
		}

		switch outExt {
		case ".yaml", ".yml":
			outData, err = yaml.Marshal(ordered)
			if err != nil {
				fmt.Printf("[ERROR] Failed to marshal to YAML: %v\n", err)
				os.Exit(1)
			}
		case ".json":
			outData, err = json.MarshalIndent(ordered, "", "  ")
			if err != nil {
				fmt.Printf("[ERROR] Failed to marshal to JSON: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Println("[ERROR] Unsupported output file format. Use .yaml/.yml or .json")
			os.Exit(1)
		}

		// Ensure output directory exists
		outDir := filepath.Dir(outputFile)
		if _, err := os.Stat(outDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Printf("[ERROR] Failed to create output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Write output
		if err := os.WriteFile(outputFile, outData, 0644); err != nil {
			fmt.Printf("[ERROR] Failed to write output file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully converted '%s' → '%s'\n", inputFile, outputFile)
	},
}

// Recursive cleanup function
func removeEmptyFields(i interface{}) interface{} {
	switch v := i.(type) {
	case map[string]interface{}:
		clean := make(map[string]interface{})
		for key, val := range v {
			val = removeEmptyFields(val)
			if val != nil {
				clean[key] = val
			}
		}
		if len(clean) == 0 {
			return nil
		}
		return clean
	case []interface{}:
		var clean []interface{}
		for _, val := range v {
			val = removeEmptyFields(val)
			if val != nil {
				clean = append(clean, val)
			}
		}
		if len(clean) == 0 {
			return nil
		}
		return clean
	case string:
		if v == "" {
			return nil
		}
		return v
	case nil:
		return nil
	default:
		return v
	}
}

var inputFile string
var outputFile string

func init() {
	convertCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input config file (yaml/json)")
	convertCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output config file (yaml/json)")
}
