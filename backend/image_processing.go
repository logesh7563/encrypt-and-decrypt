package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// FlipVertical flips an image upside-down
func FlipVertical(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	flipped := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			flipped.Set(x, height-y-1, img.At(x, y))
		}
	}
	return flipped
}

// RotateArbitrary rotates an image by the specified angle (in degrees)
func RotateArbitrary(img image.Image, angle float64) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	centerX, centerY := float64(width)/2, float64(height)/2

	// Convert angle to radians
	angleRad := angle * math.Pi / 180

	// Calculate new image dimensions to fit the rotated image
	cosAngle, sinAngle := math.Abs(math.Cos(angleRad)), math.Abs(math.Sin(angleRad))
	newWidth := int(math.Ceil(float64(width)*cosAngle + float64(height)*sinAngle))
	newHeight := int(math.Ceil(float64(width)*sinAngle + float64(height)*cosAngle))

	// Create new image with adjusted dimensions
	rotated := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Fill with transparent color
	draw.Draw(rotated, rotated.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Calculate new center
	newCenterX, newCenterY := float64(newWidth)/2, float64(newHeight)/2

	// Perform rotation
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Translate to origin
			xt := float64(x) - newCenterX
			yt := float64(y) - newCenterY

			// Rotate
			cosA, sinA := math.Cos(-angleRad), math.Sin(-angleRad)
			xr := xt*cosA - yt*sinA
			yr := xt*sinA + yt*cosA

			// Translate back and adjust for original center
			xOriginal := int(math.Round(xr + centerX))
			yOriginal := int(math.Round(yr + centerY))

			// Check if the point is in the original image
			if xOriginal >= 0 && xOriginal < width && yOriginal >= 0 && yOriginal < height {
				rotated.Set(x, y, img.At(xOriginal, yOriginal))
			}
		}
	}

	return rotated
}

// RotateShear rotates an image using three shear matrices
func RotateShear(img image.Image, angle float64) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert angle to radians
	angleRad := angle * math.Pi / 180

	// Calculate shear factors
	tanHalfAngle := math.Tan(angleRad / 2)

	// Calculate new dimensions to fit the rotated image
	cosAngle, sinAngle := math.Abs(math.Cos(angleRad)), math.Abs(math.Sin(angleRad))
	newWidth := int(math.Ceil(float64(width)*cosAngle + float64(height)*sinAngle))
	newHeight := int(math.Ceil(float64(width)*sinAngle + float64(height)*cosAngle))

	// Create intermediate and result images
	intermediate1 := image.NewRGBA(image.Rect(0, 0, width, newHeight))
	intermediate2 := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Fill with background color
	draw.Draw(intermediate1, intermediate1.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(intermediate2, intermediate2.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(result, result.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Calculate offsets for centered rotation
	offsetX := (newWidth - width) / 2
	offsetY := (newHeight - height) / 2

	// Step 1: Horizontal shear
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			newX := int(float64(x)-float64(y-height/2)*tanHalfAngle) + offsetX
			newY := y + offsetY
			if newX >= 0 && newX < newWidth && newY >= 0 && newY < newHeight {
				intermediate1.Set(newX, newY, img.At(x, y))
			}
		}
	}

	// Step 2: Vertical shear
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			newX := x
			newY := int(float64(y) + float64(x-newWidth/2)*sinAngle)
			if newX >= 0 && newX < newWidth && newY >= 0 && newY < newHeight {
				intermediate2.Set(newX, newY, intermediate1.At(x, y))
			}
		}
	}

	// Step 3: Horizontal shear again
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			newX := int(float64(x) - float64(y-newHeight/2)*tanHalfAngle)
			newY := y
			if newX >= 0 && newX < newWidth && newY >= 0 && newY < newHeight {
				result.Set(newX, newY, intermediate2.At(x, y))
			}
		}
	}

	return result
}

// ConvertToGrayscale converts an image to grayscale
func ConvertToGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}

	return grayImg
}

// ApplyBoxBlur applies a box blur to an image
func ApplyBoxBlur(img image.Image, radius int) image.Image {
	bounds := img.Bounds()
	blurred := image.NewRGBA(bounds)

	// Create kernel size based on radius
	kernelSize := 2*radius + 1
	kernelArea := float64(kernelSize * kernelSize)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a float64

			// Apply kernel
			for ky := -radius; ky <= radius; ky++ {
				for kx := -radius; kx <= radius; kx++ {
					sampleX, sampleY := x+kx, y+ky

					// Check bounds
					if sampleX < bounds.Min.X {
						sampleX = bounds.Min.X
					} else if sampleX >= bounds.Max.X {
						sampleX = bounds.Max.X - 1
					}

					if sampleY < bounds.Min.Y {
						sampleY = bounds.Min.Y
					} else if sampleY >= bounds.Max.Y {
						sampleY = bounds.Max.Y - 1
					}

					// Get pixel color
					pixelColor := img.At(sampleX, sampleY)
					rVal, gVal, bVal, aVal := pixelColor.RGBA()

					// Accumulate values (normalize from uint32 to float64)
					r += float64(rVal) / 0xffff
					g += float64(gVal) / 0xffff
					b += float64(bVal) / 0xffff
					a += float64(aVal) / 0xffff
				}
			}

			// Calculate average
			r /= kernelArea
			g /= kernelArea
			b /= kernelArea
			a /= kernelArea

			// Set pixel
			blurred.Set(x, y, color.RGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: uint8(a * 255),
			})
		}
	}

	return blurred
}

// ApplyGaussianBlur applies a Gaussian blur to an image
func ApplyGaussianBlur(img image.Image, radius float64) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	blurred := image.NewRGBA(bounds)

	// Create kernel size based on radius (typically 3σ rule)
	kernelSize := int(math.Ceil(radius*3))*2 + 1
	kernelRadius := kernelSize / 2

	// Generate 1D Gaussian kernel
	kernel := make([]float64, kernelSize)
	kernelSum := 0.0

	for i := 0; i < kernelSize; i++ {
		x := float64(i - kernelRadius)
		kernel[i] = math.Exp(-(x * x) / (2 * radius * radius))
		kernelSum += kernel[i]
	}

	// Normalize kernel
	for i := 0; i < kernelSize; i++ {
		kernel[i] /= kernelSum
	}

	// Create temporary image for horizontal pass
	tempImg := image.NewRGBA(bounds)

	// Horizontal pass
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64

			for kx := 0; kx < kernelSize; kx++ {
				sampleX := x + (kx - kernelRadius)

				// Handle edge cases
				if sampleX < 0 {
					sampleX = 0
				} else if sampleX >= width {
					sampleX = width - 1
				}

				pixelColor := img.At(sampleX, y)
				rVal, gVal, bVal, aVal := pixelColor.RGBA()

				weight := kernel[kx]
				r += float64(rVal) / 0xffff * weight
				g += float64(gVal) / 0xffff * weight
				b += float64(bVal) / 0xffff * weight
				a += float64(aVal) / 0xffff * weight
			}

			tempImg.Set(x, y, color.RGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: uint8(a * 255),
			})
		}
	}

	// Vertical pass
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64

			for ky := 0; ky < kernelSize; ky++ {
				sampleY := y + (ky - kernelRadius)

				// Handle edge cases
				if sampleY < 0 {
					sampleY = 0
				} else if sampleY >= height {
					sampleY = height - 1
				}

				pixelColor := tempImg.At(x, sampleY)
				rVal, gVal, bVal, aVal := pixelColor.RGBA()

				weight := kernel[ky]
				r += float64(rVal) / 0xffff * weight
				g += float64(gVal) / 0xffff * weight
				b += float64(bVal) / 0xffff * weight
				a += float64(aVal) / 0xffff * weight
			}

			blurred.Set(x, y, color.RGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: uint8(a * 255),
			})
		}
	}

	return blurred
}

// ApplySobelEdgeDetection applies Sobel edge detection to an image
func ApplySobelEdgeDetection(img image.Image) image.Image {
	// First convert to grayscale for edge detection
	grayImg := ConvertToGrayscale(img)
	bounds := grayImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	edges := image.NewRGBA(bounds)

	// Sobel operators
	sobelX := [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}

	sobelY := [][]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// Apply Sobel operator
			var gx, gy float64

			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					pixel := grayImg.At(x+j, y+i)
					grayValue := color.GrayModel.Convert(pixel).(color.Gray).Y

					gx += float64(grayValue) * float64(sobelX[i+1][j+1])
					gy += float64(grayValue) * float64(sobelY[i+1][j+1])
				}
			}

			// Calculate gradient magnitude
			magnitude := math.Sqrt(gx*gx + gy*gy)

			// Normalize and threshold
			normalizedMagnitude := math.Min(255, magnitude)

			// Set edge pixel
			edges.Set(x, y, color.RGBA{
				R: uint8(normalizedMagnitude),
				G: uint8(normalizedMagnitude),
				B: uint8(normalizedMagnitude),
				A: 255,
			})
		}
	}

	return edges
}
