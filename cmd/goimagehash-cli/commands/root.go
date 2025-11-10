package commands

import (
	"github.com/spf13/cobra"
)

var (
	hashType  string
	threshold int
	verbose   bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "goimagehash-cli",
	Short: "A CLI tool for image perceptual hashing",
	Long: `goimagehash-cli is a command-line interface for computing and comparing 
image hashes using various perceptual hashing algorithms including 
Average Hash, Difference Hash, Perception Hash, and more.`,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&hashType, "hash-type", "t", "average", "Hash algorithm (average, difference, perception, wavelet, double-gradient)")
	RootCmd.PersistentFlags().IntVarP(&threshold, "threshold", "x", 10, "Similarity threshold for comparisons")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	RootCmd.AddCommand(hashCmd)
	RootCmd.AddCommand(compareCmd)
	RootCmd.AddCommand(batchCmd)
}

func init() {
	cobra.EnablePrefixMatching = true
}