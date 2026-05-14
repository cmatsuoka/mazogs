package graphics

import "testing"

func pixelByteIndex(row, col int) int {
	return (row*8 + col) * 3
}

func pixelIsBlack(pixels [192]byte, row, col int) bool {
	index := pixelByteIndex(row, col)
	return pixels[index] == 0x00 && pixels[index+1] == 0x00 && pixels[index+2] == 0x00
}

func pixelIsWhite(pixels [192]byte, row, col int) bool {
	index := pixelByteIndex(row, col)
	return pixels[index] == 0xff && pixels[index+1] == 0xff && pixels[index+2] == 0xff
}

// TestBuildCharPixelsBlankGlyphIsAllWhite checks that a glyph with no set bits
// renders to an all-white 8x8 pixel buffer.
func TestBuildCharPixelsBlankGlyphIsAllWhite(t *testing.T) {
	var pixels [192]byte

	buildCharPixels([8]byte{}, &pixels)

	for i, b := range pixels {
		if b != 0xff {
			t.Fatalf("pixels[%d] = %#x, want %#x", i, b, byte(0xff))
		}
	}
}

// TestBuildCharPixelsMapsBitPositions checks that set glyph bits map to black
// pixels at the expected 8x8 coordinates.
func TestBuildCharPixelsMapsBitPositions(t *testing.T) {
	var glyph [8]byte
	glyph[0] = 0x80 // top-left pixel
	glyph[7] = 0x01 // bottom-right pixel

	var pixels [192]byte
	buildCharPixels(glyph, &pixels)

	if !pixelIsBlack(pixels, 0, 0) {
		t.Fatalf("expected top-left pixel to be black")
	}
	if !pixelIsBlack(pixels, 7, 7) {
		t.Fatalf("expected bottom-right pixel to be black")
	}
	if !pixelIsWhite(pixels, 0, 1) {
		t.Fatalf("expected neighbor of top-left pixel to remain white")
	}
	if !pixelIsWhite(pixels, 7, 6) {
		t.Fatalf("expected neighbor of bottom-right pixel to remain white")
	}
}
