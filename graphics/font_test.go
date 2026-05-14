package graphics

import "testing"

// TestFontInverseGlyphPairs checks that the upper half of the font table is
// the bitwise inverse of the lower half, as generated in init().
func TestFontInverseGlyphPairs(t *testing.T) {
	for i := 0; i < 64; i++ {
		for j := 0; j < 8; j++ {
			want := ^font[i][j]
			if got := font[i+64][j]; got != want {
				t.Fatalf("font[%d][%d] = %#x, want inverse %#x", i+64, j, got, want)
			}
		}
	}
}
