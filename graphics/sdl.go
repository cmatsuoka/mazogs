package graphics

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	width     int
	height    int
	window    *sdl.Window
	renderer  *sdl.Renderer
	image     *sdl.Surface
	screen    *sdl.Texture
	fontAtlas *sdl.Texture

	keyPressed bool
	keyValue   string
	keyLatch   string // set on KEYDOWN, cleared only when consumed by InKey
)

func Init(title string, width, height int32) error {
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var err error
	window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		return fmt.Errorf("can't create window: %s", err)
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		return fmt.Errorf("can't create renderer: %s", err)
	}

	screen, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET, 256, 192)
	if err != nil {
		return fmt.Errorf("can't create screen texture: %s", err)
	}
	screen.SetBlendMode(sdl.BLENDMODE_NONE)

	fontAtlas, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET, 128, 64)
	if err != nil {
		return fmt.Errorf("can't create fontAtlas texture: %s", err)
	}

	if err := renderer.SetRenderTarget(screen); err != nil {
		return fmt.Errorf("can't set render target: %s", err)
	}

	if err := renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff); err != nil {
		return fmt.Errorf("can't set draw color: %s", err)
	}

	buildFontAtlas()

	renderer.Clear()

	return nil
}

func Deinit() {
	if screen != nil {
		screen.Destroy()
	}
	if fontAtlas != nil {
		fontAtlas.Destroy()
	}
	if window != nil {
		window.Destroy()
	}
}

func Present() error {
	if err := renderer.SetRenderTarget(nil); err != nil {
		return err
	}
	if err := renderer.Copy(screen, nil, nil); err != nil {
		return err
	}
	renderer.Present()
	renderer.SetRenderTarget(screen)

	return nil
}

func ProcessEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.WindowEvent:
			switch event.(*sdl.WindowEvent).Event {
			case sdl.WINDOWEVENT_CLOSE:
				os.Exit(0)
			}

		case *sdl.KeyboardEvent:
			kev := event.(*sdl.KeyboardEvent)
			if kev.Repeat != 0 {
				break
			}
			if kev.Type == sdl.KEYDOWN {
				keyPressed = true
				keyValue = string(kev.Keysym.Sym)
				keyLatch = keyValue
			} else {
				// Only clear keyValue if the released key is the one
				// currently tracked. This prevents an unrelated key
				// release (modifier tap, X11 repeat artifact) from
				// interrupting a held movement key — matching the
				// ZX-81 LAST_K behaviour which always reflects the
				// physical keyboard state.
				if keyValue == string(kev.Keysym.Sym) {
					keyValue = ""
				}
			}
		}
	}
}

func WaitKey() {
	keyPressed = false
	for !keyPressed {
		ProcessEvents()
	}
}

func renderChar(row, col int, c byte) {
	if c > 64 {
		c = 64 + (c & 0x7f)
	}
	src := sdl.Rect{X: int32((c & 0x0f) << 3), Y: int32((c >> 4) << 3), W: 8, H: 8}
	dst := sdl.Rect{X: int32(col << 3), Y: int32(row << 3), W: 8, H: 8}
	renderer.Copy(fontAtlas, &src, &dst)
}

func buildFontAtlas() {
	var pixels [192]byte
	for i := byte(0); i < 128; i++ {
		dst := sdl.Rect{X: int32((i & 0x0f) << 3), Y: int32((i >> 4) << 3), W: 8, H: 8}
		buildCharPixels(font[i], &pixels)
		fontAtlas.Update(&dst, unsafe.Pointer(&pixels[0]), 8*3)
	}
}

func buildCharPixels(c [8]byte, pixels *[192]byte) {
	// initialize each character all white
	for i := 0; i < len(pixels); i++ {
		pixels[i] = 0xff
	}

	// each character definition byte
	for i := 0; i < 8; i++ {
		// each bit in a character definition byte
		mask := byte(0x80)
		for j := 0; j < 8; j++ {
			if c[i]&mask != 0 {
				index := (i<<3 + j) * 3
				pixels[index] = 0
				pixels[index+1] = 0
				pixels[index+2] = 0
			}
			mask >>= 1
		}
	}
}
