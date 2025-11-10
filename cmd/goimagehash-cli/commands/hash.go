package commands

import (
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/corona10/goimagehash"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	bits         int
)

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:   "hash [image_file]",
	Short: "Compute perceptual hash of an image",
	Long: `Compute the perceptual hash of an image using the specified algorithm.
Supported formats: JPEG, PNG, GIF, and other formats supported by Go's image package.

Examples:
  goimagehash-cli hash image.jpg
  goimagehash-cli hash -t perception image.png
  goimagehash-cli hash -t average -f hex image.jpg`,
	Args: cobra.ExactArgs(1),
	RunE: runHash,
}

func init() {
	hashCmd.Flags().StringVarP(&outputFormat, "format", "f", "binary", "Output format (binary, hex, base64)")
	hashCmd.Flags().IntVarP(&bits, "bits", "b", 64, "Hash bits (for extended hashes)")
}

func runHash(cmd *cobra.Command, args []string) error {
	imagePath := args[0]
	
	if verbose {
		fmt.Printf("Processing image: %s\n", imagePath)
		fmt.Printf("Hash algorithm: %s\n", hashType)
		fmt.Printf("Output format: %s\n", outputFormat)
	}

	// Open and decode image
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Compute hash
	var output string
	var hashKind goimagehash.Kind
	var hashBits int

	switch hashType {
	case "average", "ahash":
		hash, err := goimagehash.AverageHash(img)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}
		switch outputFormat {
		case "binary":
			output = hash.ToString()
		case "hex":
			output = fmt.Sprintf("%x", hash.GetHash())
		case "base64":
			output = hex.EncodeToString([]byte{byte(hash.GetKind())}) + fmt.Sprintf("0x%x", hash.GetHash())
		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}
		hashKind = hash.GetKind()
		hashBits = hash.Bits()

	case "difference", "dhash":
		hash, err := goimagehash.DifferenceHash(img)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}
		switch outputFormat {
		case "binary":
			output = hash.ToString()
		case "hex":
			output = fmt.Sprintf("%x", hash.GetHash())
		case "base64":
			output = hex.EncodeToString([]byte{byte(hash.GetKind())}) + fmt.Sprintf("0x%x", hash.GetHash())
		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}
		hashKind = hash.GetKind()
		hashBits = hash.Bits()

	case "perception", "phash":
		hash, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}
		switch outputFormat {
		case "binary":
			output = hash.ToString()
		case "hex":
			output = fmt.Sprintf("%x", hash.GetHash())
		case "base64":
			output = hex.EncodeToString([]byte{byte(hash.GetKind())}) + fmt.Sprintf("0x%x", hash.GetHash())
		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}
		hashKind = hash.GetKind()
		hashBits = hash.Bits()

	case "double-gradient", "dgrad":
		// For DoubleGradient, use ExtImageHash with base64 format by default
		extHash, err := goimagehash.DoubleGradientHash(img, 8, 8)
		if err != nil {
			return fmt.Errorf("failed to compute double gradient hash: %w", err)
		}
		
		switch outputFormat {
		case "binary":
			output = extHash.ToString()
		case "hex":
			output = fmt.Sprintf("%x", extHash.GetHash())
		case "base64":
			output = extHash.ToBase64()
		default:
			return fmt.Errorf("unsupported output format: %s", outputFormat)
		}
		hashKind = extHash.GetKind()
		hashBits = extHash.Bits()

	default:
		return fmt.Errorf("unsupported hash type: %s. Use: average, difference, perception, double-gradient", hashType)
	}

	fmt.Println(output)

	if verbose {
		fmt.Printf("Hash type: %v\n", hashKind)
		fmt.Printf("Bits: %d\n", hashBits)
		fmt.Printf("File: %s\n", filepath.Base(imagePath))
	}

	return nil
}