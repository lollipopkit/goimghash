package goimagehash

import (
	"encoding/base64"
	"image"

	"github.com/nfnt/resize"
)

// DoubleGradientHash implements the DoubleGradient algorithm similar to the Rust img_hash library
// DoubleGradient resizes the grayscaled image to (width/2 + 1) x (height/2 + 1) and compares 
// columns in addition to rows, combining both horizontal and vertical gradient comparisons.
func DoubleGradientHash(img image.Image, width, height int) (*ExtImageHash, error) {
	// Round dimensions to next multiple of 2 (required by DoubleGradient)
	width = int(nextMultipleOf2(uint(width)))
	height = int(nextMultipleOf2(uint(height)))
	
	// Calculate resize dimensions: (width/2 + 1) x (height/2 + 1)
	resizeWidth := width/2 + 1
	resizeHeight := height/2 + 1

	// Convert to grayscale and resize
	grayImg := image.NewGray(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	// Resize the image using Lanczos3 filter (default in Rust library)
	resized := resize.Resize(uint(resizeWidth), uint(resizeHeight), grayImg, resize.Lanczos3)

	// Extract pixel values directly from grayscale image
	pixels := make([]uint8, resizeWidth*resizeHeight)
	for y := 0; y < resizeHeight; y++ {
		for x := 0; x < resizeWidth; x++ {
			idx := y*resizeWidth + x
			// resized is already grayscale, just get the Y value
			pixel := resized.At(int(x), int(y))
			r, _, _, _ := pixel.RGBA()
			pixels[idx] = uint8(r >> 8)
		}
	}

	// Compute hash bits using DoubleGradient algorithm
	hashBits := doubleGradientHashBits(pixels, resizeWidth, resizeHeight)
	
	// Convert bits to ExtImageHash
	hashBytes := bitsToBytes(hashBits)
	totalBits := len(hashBits)
	
	return NewExtImageHash(hashBytes, DGHash, totalBits), nil
}

// nextMultipleOf2 rounds up to the next multiple of 2
func nextMultipleOf2(x uint) uint {
	if x%2 == 0 {
		return x
	}
	return x + 1
}

// doubleGradientHashBits implements the core DoubleGradient algorithm
// Combines horizontal gradient (rows) and vertical gradient (columns) comparisons
func doubleGradientHashBits(pixels []uint8, width, height int) []bool {
	rowstride := width
	totalBits := width*height + width*height // horizontal + vertical comparisons
	hashBits := make([]bool, 0, totalBits)

	// Horizontal gradient comparisons (like Gradient algorithm)
	for y := 0; y < height; y++ {
		rowStart := y * rowstride
		for x := 1; x < width; x++ {
			current := pixels[rowStart+x]
			previous := pixels[rowStart+x-1]
			hashBits = append(hashBits, current > previous)
		}
	}

	// Vertical gradient comparisons (like VertGradient algorithm)  
	for x := 0; x < width; x++ {
		for y := 1; y < height; y++ {
			current := pixels[y*rowstride+x]
			previous := pixels[(y-1)*rowstride+x]
			hashBits = append(hashBits, current > previous)
		}
	}

	return hashBits
}

// bitsToBytes converts boolean bits to bytes (LSB first - matching Rust library)
func bitsToBytes(bits []bool) []uint64 {
	var result []uint64
	var currentWord uint64
	var bitPos uint

	for _, bit := range bits {
		if bit {
			currentWord |= 1 << bitPos
		}
		bitPos++
		
		if bitPos >= 64 {
			result = append(result, currentWord)
			currentWord = 0
			bitPos = 0
		}
	}
	
	if bitPos > 0 {
		result = append(result, currentWord)
	}
	
	return result
}

// ToBase64 converts ExtImageHash to base64 string without padding (like Rust library)
func (h *ExtImageHash) ToBase64() string {
	// Convert hash to bytes using existing serialization
	hashBytes := make([]byte, 0)
	for _, hashWord := range h.hash {
		for i := uint(0); i < 8; i++ {
			hashBytes = append(hashBytes, byte(hashWord>> (i*8) & 0xFF))
		}
	}
	
	// Only keep the actual bits needed, not the full bytes
	if h.bits < len(hashBytes)*8 {
		fullBytes := h.bits / 8
		if h.bits%8 != 0 {
			fullBytes++
		}
		hashBytes = hashBytes[:fullBytes]
	}
	
	// Encode without padding
	encoded := base64.RawStdEncoding.EncodeToString(hashBytes)
	return encoded
}

// DoubleGradientHashToBase64 is a convenience function that computes DoubleGradient hash
// and returns base64 string without padding (compatible with Rust img_hash library)
func DoubleGradientHashToBase64(img image.Image, width, height int) (string, error) {
	hash, err := DoubleGradientHash(img, width, height)
	if err != nil {
		return "", err
	}
	return hash.ToBase64(), nil
}