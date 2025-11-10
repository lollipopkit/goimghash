# goimagehash-cli

A command-line interface for the goimagehash library, providing perceptual hashing capabilities for images.

## Installation

```bash
go install github.com/corona10/goimagehash/cmd/goimagehash-cli@latest
```

## Usage

### Basic Commands

#### Compute Image Hash
```bash
# Compute average hash (default)
goimagehash-cli hash image.jpg

# Use different hash algorithms
goimagehash-cli hash -t perception image.png
goimagehash-cli hash -t difference image.gif
goimagehash-cli hash -t average image.jpg

# Output in different formats
goimagehash-cli hash -f hex image.jpg
goimagehash-cli hash -f binary image.png
goimagehash-cli hash -f base64 image.jpg
```

#### Compare Two Images
```bash
# Compare two images using default average hash
goimagehash-cli compare image1.jpg image2.jpg

# Use perception hash with custom threshold
goimagehash-cli compare -t perception -x 5 image1.jpg image2.jpg

# Verbose output
goimagehash-cli compare -v image1.jpg image2.jpg
```

#### Batch Processing
```bash
# Compute hashes for all images in directory
goimagehash-cli batch ./images

# Process recursively with specific extensions
goimagehash-cli batch -r -e "jpg,png" ./photos

# Find similar images
goimagehash-cli batch -d -x 10 ./images

# Export results to CSV
goimagehash-cli batch -o results.csv ./images
goimagehash-cli batch -d -o duplicates.csv ./photos
```

### Command Reference

#### Global Options
- `-t, --hash-type`: Hash algorithm (average, difference, perception) [default: average]
- `-x, --threshold`: Similarity threshold for comparisons [default: 10]
- `-v, --verbose`: Enable verbose output

#### hash Command
Computes perceptual hash of a single image.

**Options:**
- `-f, --format`: Output format (binary, hex, base64) [default: binary]
- `-b, --bits`: Hash bits for extended hashes [default: 64]

**Examples:**
```bash
goimagehash-cli hash image.jpg
goimagehash-cli hash -t perception -f hex image.png
```

#### compare Command
Compares two images and computes similarity.

**Examples:**
```bash
goimagehash-cli compare image1.jpg image2.jpg
goimagehash-cli compare -t perception -x 5 img1.png img2.png
```

Returns:
- Exit code 0: Images are similar (distance <= threshold)
- Exit code 1: Images are different (distance > threshold)

#### batch Command
Processes multiple images in batch.

**Options:**
- `-o, --output`: Output file for results (CSV format)
- `-r, --recursive`: Process directories recursively
- `-e, --extensions`: File extensions to process [default: jpg,jpeg,png,gif]
- `-d, --duplicates`: Find duplicate/similar images instead of computing hashes

**Examples:**
```bash
# Compute hashes
goimagehash-cli batch ./images
goimagehash-cli batch -r -o hashes.csv ./photos

# Find duplicates
goimagehash-cli batch -d ./images
goimagehash-cli batch -d -x 5 -o duplicates.csv ./photos
```

### Supported Image Formats
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- And other formats supported by Go's image package

### Hash Algorithms

#### Average Hash (AHash)
- Fast and simple
- Good for finding exact duplicates
- Default algorithm used

#### Difference Hash (DHash)
- Based on adjacent pixel differences
- Faster than perception hash
- Good for finding similar images

#### Perception Hash (PHash)
- Based on DCT (Discrete Cosine Transform)
- More robust to changes
- Best for finding visually similar images

### Output Formats

#### Binary
```
1101010101010101010101010101010101010101010101010101010101010101
```

#### Hex
```
a9a9a9a9a9a9a9a9
```

#### Base64
``
q6q6q6q6q6q6q6q
```

### Scripting Examples

#### Find duplicates in photo library
```bash
#!/bin/bash
goimagehash-cli batch -d -x 5 -o duplicates.csv ~/Photos
echo "Found $(cat duplicates.csv | wc -l) duplicate groups"
```

#### Check if images are similar
```bash
#!/bin/bash
if goimagehash-cli compare -x 10 image1.jpg image2.jpg >/dev/null 2>&1; then
    echo "Images are similar"
else
    echo "Images are different"
fi
```

#### Generate hash manifest
```bash
#!/bin/bash
goimagehash-cli hash -f hex *.jpg > hashes.txt
```