package commands

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/corona10/goimagehash"
	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare [image1] [image2]",
	Short: "Compare two images and compute similarity",
	Long: `Compare two images by computing their perceptual hashes and calculating
the Hamming distance between them. Lower distance values indicate higher similarity.

The command outputs the Hamming distance and whether the images are considered 
similar based on the threshold.

Examples:
  goimagehash-cli compare image1.jpg image2.jpg
  goimagehash-cli compare -t perception -x 5 img1.png img2.png`,
	Args: cobra.ExactArgs(2),
	RunE: runCompare,
}

func runCompare(cmd *cobra.Command, args []string) error {
	image1Path := args[0]
	image2Path := args[1]

	if verbose {
		fmt.Printf("Comparing images:\n")
		fmt.Printf("  Image 1: %s\n", image1Path)
		fmt.Printf("  Image 2: %s\n", image2Path)
		fmt.Printf("  Hash algorithm: %s\n", hashType)
		fmt.Printf("  Similarity threshold: %d\n", threshold)
	}

	// Load and decode first image
	img1, err := loadImage(image1Path)
	if err != nil {
		return fmt.Errorf("failed to load first image: %w", err)
	}

	// Load and decode second image
	img2, err := loadImage(image2Path)
	if err != nil {
		return fmt.Errorf("failed to load second image: %w", err)
	}

	// Compute hashes
	var distance int
	var similar bool
	var hashKind1, hashKind2 goimagehash.Kind
	var hashStr1, hashStr2 string

	switch hashType {
	case "double-gradient", "dgrad":
		// Handle DoubleGradient with ExtImageHash
		extHash1, err := goimagehash.DoubleGradientHash(img1, 8, 8)
		if err != nil {
			return fmt.Errorf("failed to compute hash for first image: %w", err)
		}

		extHash2, err := goimagehash.DoubleGradientHash(img2, 8, 8)
		if err != nil {
			return fmt.Errorf("failed to compute hash for second image: %w", err)
		}

		// Calculate distance
		distance, err = extHash1.Distance(extHash2)
		if err != nil {
			return fmt.Errorf("failed to calculate distance: %w", err)
		}

		similar = distance <= threshold
		hashKind1 = extHash1.GetKind()
		hashKind2 = extHash2.GetKind()
		hashStr1 = extHash1.ToString()
		hashStr2 = extHash2.ToString()

	default:
		// Handle standard ImageHash types
		hash1, err := computeHash(img1)
		if err != nil {
			return fmt.Errorf("failed to compute hash for first image: %w", err)
		}

		hash2, err := computeHash(img2)
		if err != nil {
			return fmt.Errorf("failed to compute hash for second image: %w", err)
		}

		// Calculate distance
		distance, err = hash1.Distance(hash2)
		if err != nil {
			return fmt.Errorf("failed to calculate distance: %w", err)
		}

		similar = distance <= threshold
		hashKind1 = hash1.GetKind()
		hashKind2 = hash2.GetKind()
		hashStr1 = hash1.ToString()
		hashStr2 = hash2.ToString()
	}

	// Output results
	status := "different"
	if similar {
		status = "similar"
	}

	fmt.Printf("Distance: %d\n", distance)
	fmt.Printf("Status: %s (threshold: %d)\n", status, threshold)

	if verbose {
		fmt.Printf("Hash 1: %s\n", hashStr1)
		fmt.Printf("Hash 2: %s\n", hashStr2)
		fmt.Printf("Hash 1 type: %v\n", hashKind1)
		fmt.Printf("Hash 2 type: %v\n", hashKind2)
	}

	// Set exit code for scripting
	if !similar {
		os.Exit(1)
	}

	return nil
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func computeHash(img image.Image) (*goimagehash.ImageHash, error) {
	switch hashType {
	case "average", "ahash":
		return goimagehash.AverageHash(img)
	case "difference", "dhash":
		return goimagehash.DifferenceHash(img)
	case "perception", "phash":
		return goimagehash.PerceptionHash(img)
	default:
		return nil, fmt.Errorf("unsupported hash type: %s", hashType)
	}
}