package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// FlipUpsideDown flips an image upside-down (vertical flip)
func FlipUpsideDown(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	flipped := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get the pixel color
			srcColor := img.At(x, y)
			r, g, b, a := srcColor.RGBA()

			// Set the pixel in the flipped position
			destY := height - y - 1
			flipped.Set(x, destY, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}
	return flipped
}

// FlipHorizontal flips an image horizontally (left to right)
func FlipHorizontal(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	flipped := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get the pixel color
			srcColor := img.At(x, y)
			r, g, b, a := srcColor.RGBA()

			// Set the pixel in the flipped position
			destX := width - x - 1
			flipped.Set(destX, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
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

	// Get shear factors for 3-shear rotation
	alpha := -math.Tan(angleRad / 2)
	beta := math.Sin(angleRad)

	// Calculate new dimensions to fit the rotated image
	cosAngle, sinAngle := math.Abs(math.Cos(angleRad)), math.Abs(math.Sin(angleRad))
	newWidth := int(math.Ceil(float64(width)*cosAngle + float64(height)*sinAngle))
	newHeight := int(math.Ceil(float64(height)*cosAngle + float64(width)*sinAngle))

	// Create intermediate images
	intermediate1 := image.NewRGBA(image.Rect(0, 0, newWidth, height))
	intermediate2 := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Fill with transparent background
	draw.Draw(intermediate1, intermediate1.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(intermediate2, intermediate2.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(result, result.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Calculate centering offsets
	xOffset := (newWidth - width) / 2
	yOffset := (newHeight - height) / 2

	// Step 1: Horizontal shear
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Apply first shear: x' = x + alpha * (y - height/2)
			dy := y - height/2
			newX := int(math.Round(float64(x)+alpha*float64(dy))) + xOffset

			if newX >= 0 && newX < newWidth {
				intermediate1.Set(newX, y, img.At(x, y))
			}
		}
	}

	// Step 2: Vertical shear
	for y := 0; y < height; y++ {
		for x := 0; x < newWidth; x++ {
			// Apply second shear: y' = y + beta * (x - newWidth/2)
			dx := x - newWidth/2
			newY := int(math.Round(float64(y)+beta*float64(dx))) + yOffset

			if newY >= 0 && newY < newHeight {
				intermediate2.Set(x, newY, intermediate1.At(x, y))
			}
		}
	}

	// Step 3: Horizontal shear again
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Apply third shear: x' = x + alpha * (y - newHeight/2)
			dy := y - newHeight/2
			newX := int(math.Round(float64(x) + alpha*float64(dy)))

			if newX >= 0 && newX < newWidth {
				result.Set(newX, y, intermediate2.At(x, y))
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
			// Get RGB values
			pixelColor := img.At(x, y)
			r, g, b, _ := pixelColor.RGBA()

			// Convert to grayscale using luminance formula
			// Y = 0.299*R + 0.587*G + 0.114*B
			luma := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 0xffff

			// Set the grayscale pixel
			grayImg.SetGray(x, y, color.Gray{Y: uint8(luma * 255)})
		}
	}

	return grayImg
}

// ApplyBoxBlur applies a box blur to an image
func ApplyBoxBlur(img image.Image, radius int) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	blurred := image.NewRGBA(bounds)

	// Ensure radius is at least 1
	if radius < 1 {
		radius = 1
	}

	// Create kernel size based on radius
	kernelSize := 2*radius + 1

	// Box blur is separable, so we'll do horizontal and vertical passes
	// Create temporary image for horizontal pass
	tempImg := image.NewRGBA(bounds)

	// Horizontal pass
	for y := 0; y < height; y++ {
		// Sliding window approach for better performance
		var sumR, sumG, sumB, sumA float64

		// Initialize the sliding window for the first pixel in the row
		for kx := -radius; kx <= radius; kx++ {
			sampleX := kx
			// Edge handling - mirror edge pixels
			if sampleX < 0 {
				sampleX = -sampleX
			} else if sampleX >= width {
				sampleX = 2*width - sampleX - 1
			}

			pixelColor := img.At(sampleX, y)
			r, g, b, a := pixelColor.RGBA()
			sumR += float64(r) / 0xffff
			sumG += float64(g) / 0xffff
			sumB += float64(b) / 0xffff
			sumA += float64(a) / 0xffff
		}

		// Set the first pixel
		tempImg.Set(0, y, color.RGBA{
			R: uint8(sumR / float64(kernelSize) * 255),
			G: uint8(sumG / float64(kernelSize) * 255),
			B: uint8(sumB / float64(kernelSize) * 255),
			A: uint8(sumA / float64(kernelSize) * 255),
		})

		// Process the rest of the row using sliding window
		for x := 1; x < width; x++ {
			// Remove leftmost pixel from the sums
			leftX := x - radius - 1
			if leftX < 0 {
				leftX = -leftX
			} else if leftX >= width {
				leftX = 2*width - leftX - 1
			}
			leftPixel := img.At(leftX, y)
			lr, lg, lb, la := leftPixel.RGBA()
			sumR -= float64(lr) / 0xffff
			sumG -= float64(lg) / 0xffff
			sumB -= float64(lb) / 0xffff
			sumA -= float64(la) / 0xffff

			// Add rightmost pixel to the sums
			rightX := x + radius
			if rightX < 0 {
				rightX = -rightX
			} else if rightX >= width {
				rightX = 2*width - rightX - 1
			}
			rightPixel := img.At(rightX, y)
			rr, rg, rb, ra := rightPixel.RGBA()
			sumR += float64(rr) / 0xffff
			sumG += float64(rg) / 0xffff
			sumB += float64(rb) / 0xffff
			sumA += float64(ra) / 0xffff

			// Set the current pixel
			tempImg.Set(x, y, color.RGBA{
				R: uint8(sumR / float64(kernelSize) * 255),
				G: uint8(sumG / float64(kernelSize) * 255),
				B: uint8(sumB / float64(kernelSize) * 255),
				A: uint8(sumA / float64(kernelSize) * 255),
			})
		}
	}

	// Vertical pass
	for x := 0; x < width; x++ {
		// Sliding window approach for the vertical pass
		var sumR, sumG, sumB, sumA float64

		// Initialize the sliding window for the first pixel in the column
		for ky := -radius; ky <= radius; ky++ {
			sampleY := ky
			// Edge handling - mirror edge pixels
			if sampleY < 0 {
				sampleY = -sampleY
			} else if sampleY >= height {
				sampleY = 2*height - sampleY - 1
			}

			pixelColor := tempImg.At(x, sampleY)
			r, g, b, a := pixelColor.RGBA()
			sumR += float64(r) / 0xffff
			sumG += float64(g) / 0xffff
			sumB += float64(b) / 0xffff
			sumA += float64(a) / 0xffff
		}

		// Set the first pixel
		blurred.Set(x, 0, color.RGBA{
			R: uint8(sumR / float64(kernelSize) * 255),
			G: uint8(sumG / float64(kernelSize) * 255),
			B: uint8(sumB / float64(kernelSize) * 255),
			A: uint8(sumA / float64(kernelSize) * 255),
		})

		// Process the rest of the column using sliding window
		for y := 1; y < height; y++ {
			// Remove topmost pixel from the sums
			topY := y - radius - 1
			if topY < 0 {
				topY = -topY
			} else if topY >= height {
				topY = 2*height - topY - 1
			}
			topPixel := tempImg.At(x, topY)
			tr, tg, tb, ta := topPixel.RGBA()
			sumR -= float64(tr) / 0xffff
			sumG -= float64(tg) / 0xffff
			sumB -= float64(tb) / 0xffff
			sumA -= float64(ta) / 0xffff

			// Add bottommost pixel to the sums
			bottomY := y + radius
			if bottomY < 0 {
				bottomY = -bottomY
			} else if bottomY >= height {
				bottomY = 2*height - bottomY - 1
			}
			bottomPixel := tempImg.At(x, bottomY)
			br, bg, bb, ba := bottomPixel.RGBA()
			sumR += float64(br) / 0xffff
			sumG += float64(bg) / 0xffff
			sumB += float64(bb) / 0xffff
			sumA += float64(ba) / 0xffff

			// Set the current pixel
			blurred.Set(x, y, color.RGBA{
				R: uint8(sumR / float64(kernelSize) * 255),
				G: uint8(sumG / float64(kernelSize) * 255),
				B: uint8(sumB / float64(kernelSize) * 255),
				A: uint8(sumA / float64(kernelSize) * 255),
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

	// Make sure radius is at least 0.1 to avoid division by zero
	if radius < 0.1 {
		radius = 0.1
	}

	// Create kernel size based on radius (typically 3σ rule)
	// Ensure kernel size is odd for symmetric processing
	kernelSize := int(math.Ceil(radius*3))*2 + 1
	kernelRadius := kernelSize / 2

	// Generate 1D Gaussian kernel
	kernel := make([]float64, kernelSize)
	kernelSum := 0.0

	for i := 0; i < kernelSize; i++ {
		x := float64(i - kernelRadius)
		// Gaussian function: exp(-x²/2σ²)
		kernel[i] = math.Exp(-(x * x) / (2 * radius * radius))
		kernelSum += kernel[i]
	}

	// Normalize kernel to ensure weights sum to 1
	for i := 0; i < kernelSize; i++ {
		kernel[i] /= kernelSum
	}

	// Create temporary image for separable filtering
	tempImg := image.NewRGBA(bounds)

	// Horizontal pass with better edge handling
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64

			for kx := 0; kx < kernelSize; kx++ {
				// Calculate sample position with proper kernel offset
				sampleX := x + (kx - kernelRadius)

				// Edge handling - mirror edge pixels
				if sampleX < 0 {
					sampleX = -sampleX
				} else if sampleX >= width {
					sampleX = 2*width - sampleX - 1
				}

				// Get pixel color at the sample position
				pixelColor := img.At(sampleX, y)
				rVal, gVal, bVal, aVal := pixelColor.RGBA()

				// Apply weight from kernel
				weight := kernel[kx]
				r += float64(rVal) / 0xffff * weight
				g += float64(gVal) / 0xffff * weight
				b += float64(bVal) / 0xffff * weight
				a += float64(aVal) / 0xffff * weight
			}

			// Set the blurred pixel in the temporary image
			tempImg.Set(x, y, color.RGBA{
				R: uint8(r * 255),
				G: uint8(g * 255),
				B: uint8(b * 255),
				A: uint8(a * 255),
			})
		}
	}

	// Vertical pass with better edge handling
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64

			for ky := 0; ky < kernelSize; ky++ {
				// Calculate sample position with proper kernel offset
				sampleY := y + (ky - kernelRadius)

				// Edge handling - mirror edge pixels
				if sampleY < 0 {
					sampleY = -sampleY
				} else if sampleY >= height {
					sampleY = 2*height - sampleY - 1
				}

				// Get pixel color at the sample position
				pixelColor := tempImg.At(x, sampleY)
				rVal, gVal, bVal, aVal := pixelColor.RGBA()

				// Apply weight from kernel
				weight := kernel[ky]
				r += float64(rVal) / 0xffff * weight
				g += float64(gVal) / 0xffff * weight
				b += float64(bVal) / 0xffff * weight
				a += float64(aVal) / 0xffff * weight
			}

			// Set the final blurred pixel
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

	// First pass to determine max gradient for normalization
	var maxGradient float64 = 0.1 // Small value to avoid division by zero

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
			if magnitude > maxGradient {
				maxGradient = magnitude
			}
		}
	}

	// Threshold for edge detection (adjust as needed)
	threshold := maxGradient * 0.2

	// Second pass to apply thresholding and create final image
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

			// Apply threshold and normalize
			var pixelValue uint8
			if magnitude < threshold {
				pixelValue = 0
			} else {
				// Normalize to [0, 255] with enhanced contrast
				normalized := math.Min(255, (magnitude/maxGradient)*255)
				pixelValue = uint8(normalized)
			}

			// Set edge pixel
			edges.Set(x, y, color.RGBA{
				R: pixelValue,
				G: pixelValue,
				B: pixelValue,
				A: 255,
			})
		}
	}

	return edges
}
