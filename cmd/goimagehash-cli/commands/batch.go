package commands

import (
	"encoding/csv"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/spf13/cobra"
)

var (
	outputFile    string
	recursive     bool
	extensions    []string
	findDuplicates bool
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch [directory]",
	Short: "Process multiple images in batch",
	Long: `Process multiple images in a directory, computing hashes for all images
or finding duplicate/similar images.

Examples:
  goimagehash-cli batch ./images
  goimagehash-cli batch -r -o hashes.csv ./photos
  goimagehash-cli batch -d -x 5 ./images`,
	Args: cobra.ExactArgs(1),
	RunE: runBatch,
}

func init() {
	batchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for results (CSV format)")
	batchCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Process directories recursively")
	batchCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{"jpg", "jpeg", "png", "gif"}, "File extensions to process")
	batchCmd.Flags().BoolVarP(&findDuplicates, "duplicates", "d", false, "Find duplicate/similar images instead of computing hashes")
}

func runBatch(cmd *cobra.Command, args []string) error {
	directory := args[0]

	if verbose {
		fmt.Printf("Processing directory: %s\n", directory)
		fmt.Printf("Recursive: %v\n", recursive)
		fmt.Printf("Extensions: %v\n", extensions)
		if findDuplicates {
			fmt.Printf("Finding duplicates with threshold: %d\n", threshold)
		}
	}

	// Find all image files
	imageFiles, err := findImageFiles(directory, recursive, extensions)
	if err != nil {
		return fmt.Errorf("failed to find image files: %w", err)
	}

	if len(imageFiles) == 0 {
		fmt.Println("No image files found")
		return nil
	}

	if verbose {
		fmt.Printf("Found %d image files\n", len(imageFiles))
	}

	if findDuplicates {
		return findSimilarImages(imageFiles)
	}

	return computeBatchHashes(imageFiles)
}

func findImageFiles(dir string, recursive bool, extensions []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && !recursive && path != dir {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

func computeBatchHashes(imageFiles []string) error {
	var records [][]string
	records = append(records, []string{"File", "Hash", "HashType", "Bits"})

	for _, imagePath := range imageFiles {
		img, err := loadImage(imagePath)
		if err != nil {
			if verbose {
				fmt.Printf("Error loading %s: %v\n", imagePath, err)
			}
			continue
		}

		var hashStr, kindStr string
		var bits int

		switch hashType {
		case "double-gradient", "dgrad":
			// Handle DoubleGradient with ExtImageHash
			extHash, err := goimagehash.DoubleGradientHash(img, 8, 8)
			if err != nil {
				if verbose {
					fmt.Printf("Error computing hash for %s: %v\n", imagePath, err)
				}
				continue
			}
			hashStr = extHash.ToString()
			kindStr = "double-gradient"
			bits = extHash.Bits()

		default:
			// Handle standard ImageHash types
			hash, err := computeHash(img)
			if err != nil {
				if verbose {
					fmt.Printf("Error computing hash for %s: %v\n", imagePath, err)
				}
				continue
			}

			hashStr = hash.ToString()
			bits = hash.Bits()
			
			kindStr = "unknown"
			switch hash.GetKind() {
			case goimagehash.AHash:
				kindStr = "average"
			case goimagehash.PHash:
				kindStr = "perception"
			case goimagehash.DHash:
				kindStr = "difference"
			case goimagehash.WHash:
				kindStr = "wavelet"
			}
		}
		
		record := []string{
			imagePath,
			hashStr,
			kindStr,
			fmt.Sprintf("%d", bits),
		}
		records = append(records, record)

		if verbose {
			fmt.Printf("Processed: %s\n", imagePath)
		}
	}

	// Output results
	if outputFile != "" {
		err := writeCSV(outputFile, records)
		if err != nil {
			return fmt.Errorf("failed to write CSV file: %w", err)
		}
		fmt.Printf("Results written to: %s\n", outputFile)
	} else {
		for _, record := range records {
			fmt.Printf("%s: %s (%s)\n", record[0], record[1], record[2])
		}
	}

	return nil
}

func findSimilarImages(imageFiles []string) error {
	type ImageInfo struct {
		Path string
		Hash *goimagehash.ImageHash
		ExtHash *goimagehash.ExtImageHash
		IsExt bool
	}

	var images []ImageInfo

	// Compute hashes for all images
	for _, imagePath := range imageFiles {
		img, err := loadImage(imagePath)
		if err != nil {
			if verbose {
				fmt.Printf("Error loading %s: %v\n", imagePath, err)
			}
			continue
		}

		switch hashType {
		case "double-gradient", "dgrad":
			// Handle DoubleGradient with ExtImageHash
			extHash, err := goimagehash.DoubleGradientHash(img, 8, 8)
			if err != nil {
				if verbose {
					fmt.Printf("Error computing hash for %s: %v\n", imagePath, err)
				}
				continue
			}
			images = append(images, ImageInfo{Path: imagePath, ExtHash: extHash, IsExt: true})

		default:
			// Handle standard ImageHash types
			hash, err := computeHash(img)
			if err != nil {
				if verbose {
					fmt.Printf("Error computing hash for %s: %v\n", imagePath, err)
				}
				continue
			}
			images = append(images, ImageInfo{Path: imagePath, Hash: hash, IsExt: false})
		}

		if verbose {
			fmt.Printf("Processed: %s\n", imagePath)
		}
	}

	// Find similar images
	var groups [][]ImageInfo
	processed := make(map[int]bool)

	for i, img1 := range images {
		if processed[i] {
			continue
		}

		var group []ImageInfo
		group = append(group, img1)
		processed[i] = true

		for j, img2 := range images {
			if i == j || processed[j] {
				continue
			}

			// Only compare same hash types
			if img1.IsExt != img2.IsExt {
				continue
			}

			var distance int
			var err error

			if img1.IsExt {
				// Compare ExtImageHash
				distance, err = img1.ExtHash.Distance(img2.ExtHash)
			} else {
				// Compare ImageHash
				distance, err = img1.Hash.Distance(img2.Hash)
			}

			if err != nil {
				continue
			}

			if distance <= threshold {
				group = append(group, img2)
				processed[j] = true
			}
		}

		if len(group) > 1 {
			groups = append(groups, group)
		}
	}

	// Output results
	if len(groups) == 0 {
		fmt.Println("No similar images found")
		return nil
	}

	fmt.Printf("Found %d groups of similar images:\n\n", len(groups))

	for i, group := range groups {
		fmt.Printf("Group %d (threshold: %d):\n", i+1, threshold)
		for _, img := range group {
			fmt.Printf("  %s\n", img.Path)
		}
		fmt.Println()
	}

	if outputFile != "" {
		var records [][]string
		records = append(records, []string{"Group", "File", "Hash"})

		for i, group := range groups {
			for _, img := range group {
				var hashStr string
				if img.IsExt {
					hashStr = img.ExtHash.ToString()
				} else {
					hashStr = img.Hash.ToString()
				}
				records = append(records, []string{
					fmt.Sprintf("Group %d", i+1),
					img.Path,
					hashStr,
				})
			}
		}

		err := writeCSV(outputFile, records)
		if err != nil {
			return fmt.Errorf("failed to write CSV file: %w", err)
		}
		fmt.Printf("Results written to: %s\n", outputFile)
	}

	return nil
}

func writeCSV(filename string, records [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(records)
}