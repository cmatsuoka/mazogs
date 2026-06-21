package graphics

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Zyko0/go-sdl3/sdl"
)

var (
	window        *sdl.Window
	renderer      *sdl.Renderer
	screen        *sdl.Texture
	fontAtlas     *sdl.Texture
	quitRequested bool
	minWindowW    int32
	minWindowH    int32

	keyPressed bool
	keyValue   string
	keyLatch   string // set on KEYDOWN, cleared only when consumed by InKey
)

const (
	internalWidth  = 256
	internalHeight = 192
	aspectRatio    = float32(internalWidth) / float32(internalHeight)
)

func Init(title string, width, height int32) error {
	minWindowW = width
	minWindowH = height

	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		return fmt.Errorf("can't load SDL library: %s", err)
	}
	defer func() {
		if window == nil {
			_ = sdl.CloseLibrary()
		}
	}()

	if err := sdl.SetAppMetadata("Mazogs", "dev", "github.com.cmatsuoka.mazogs"); err != nil {
		return fmt.Errorf("can't set SDL app metadata: %s", err)
	}

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return fmt.Errorf("can't initialize SDL video: %s", err)
	}
	defer func() {
		if window == nil {
			sdl.Quit()
		}
	}()

	var err error
	window, renderer, err = sdl.CreateWindowAndRenderer(title, int(width), int(height), sdl.WINDOW_HIGH_PIXEL_DENSITY|sdl.WINDOW_RESIZABLE)
	if err != nil {
		return fmt.Errorf("can't create window and renderer: %s", err)
	}
	defer func() {
		if screen == nil {
			if renderer != nil {
				renderer.Destroy()
				renderer = nil
			}
			if window != nil {
				window.Destroy()
				window = nil
			}
		}
	}()

	if err := renderer.SetVSync(1); err != nil {
		return fmt.Errorf("can't enable vsync: %s", err)
	}
	if err := renderer.SetLogicalPresentation(internalWidth, internalHeight, sdl.LOGICAL_PRESENTATION_LETTERBOX); err != nil {
		return fmt.Errorf("can't set logical presentation: %s", err)
	}

	if err := window.SetResizable(true); err != nil {
		return fmt.Errorf("can't make window resizable: %s", err)
	}
	if err := window.SetMinimumSize(minWindowW, minWindowH); err != nil {
		return fmt.Errorf("can't set minimum window size: %s", err)
	}
	if err := window.SetAspectRatio(aspectRatio, aspectRatio); err != nil {
		return fmt.Errorf("can't set window aspect ratio: %s", err)
	}

	screen, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET, internalWidth, internalHeight)
	if err != nil {
		return fmt.Errorf("can't create screen texture: %s", err)
	}
	_ = screen.SetBlendMode(sdl.BLENDMODE_NONE)
	_ = screen.SetScaleMode(sdl.SCALEMODE_LINEAR)

	fontAtlas, err = renderer.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_STATIC, 128, 64)
	if err != nil {
		return fmt.Errorf("can't create fontAtlas texture: %s", err)
	}
	_ = fontAtlas.SetScaleMode(sdl.SCALEMODE_LINEAR)

	if err := renderer.SetRenderTarget(screen); err != nil {
		return fmt.Errorf("can't set render target: %s", err)
	}

	if err := renderer.SetDrawColor(0xff, 0xff, 0xff, 0xff); err != nil {
		return fmt.Errorf("can't set draw color: %s", err)
	}

	buildFontAtlas()

	_ = renderer.Clear()
	quitRequested = false

	return nil
}

func Deinit() {
	if screen != nil {
		screen.Destroy()
	}
	if fontAtlas != nil {
		fontAtlas.Destroy()
	}
	if renderer != nil {
		renderer.Destroy()
	}
	if window != nil {
		window.Destroy()
		window = nil
	}
	renderer = nil
	screen = nil
	fontAtlas = nil
	quitRequested = false
	minWindowW = 0
	minWindowH = 0
	keyPressed = false
	keyValue = ""
	keyLatch = ""
	sdl.Quit()
	_ = sdl.CloseLibrary()
}

func toggleFullscreen() {
	if window == nil {
		return
	}
	_ = window.SetFullscreen(!IsFullscreen())
}

func IsFullscreen() bool {
	if window == nil {
		return false
	}
	return window.Flags()&sdl.WINDOW_FULLSCREEN != 0
}

func Present() {
	if renderer == nil || screen == nil {
		return
	}
	if err := renderer.SetRenderTarget(nil); err != nil {
		return
	}
	// Always restore the render target to screen once we have switched away
	// from it, even if Copy fails. This keeps the renderer in a consistent
	// state for subsequent Present or renderChar calls.
	defer func() { _ = renderer.SetRenderTarget(screen) }()
	if err := renderer.RenderTexture(screen, nil, nil); err != nil {
		return
	}
	_ = renderer.Present()
}

func ProcessEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		handleEvent(event)
	}
	return quitRequested
}

func isModifierKey(scancode sdl.Scancode) bool {
	switch scancode {
	case sdl.SCANCODE_LCTRL, sdl.SCANCODE_RCTRL,
		sdl.SCANCODE_LSHIFT, sdl.SCANCODE_RSHIFT,
		sdl.SCANCODE_LALT, sdl.SCANCODE_RALT,
		sdl.SCANCODE_LGUI, sdl.SCANCODE_RGUI:
		return true
	}
	return false
}

func normalizeKeyName(name string) string {
	if len(name) == 1 {
		r := rune(name[0])
		if unicode.IsLetter(r) {
			return strings.ToLower(name)
		}
	}
	return name
}

func WaitKey() string {
	keyPressed = false
	for !keyPressed {
		if quitRequested {
			return ""
		}

		var event sdl.Event
		if err := sdl.WaitEvent(&event); err != nil {
			continue
		}
		handleEvent(event)
	}
	// Return keyLatch rather than keyValue: keyLatch is set on KEYDOWN and
	// only cleared when consumed, so it holds the key even if a KEYUP event
	// arrives in the same ProcessEvents poll and clears keyValue first.
	k := keyLatch
	keyLatch = ""
	return k
}

func QuitRequested() bool {
	return quitRequested
}

func shouldQuitForWindowEvent(event sdl.Event) bool {
	if window == nil {
		return true
	}
	windowID, err := window.ID()
	if err != nil {
		return true
	}
	return event.WindowEvent().WindowID == windowID
}

func handleEvent(event sdl.Event) {
	switch event.Type {
	case sdl.EVENT_QUIT:
		quitRequested = true
	case sdl.EVENT_WINDOW_CLOSE_REQUESTED:
		if shouldQuitForWindowEvent(event) {
			quitRequested = true
		}
	case sdl.EVENT_WINDOW_EXPOSED,
		sdl.EVENT_WINDOW_RESIZED,
		sdl.EVENT_WINDOW_PIXEL_SIZE_CHANGED:
		if shouldQuitForWindowEvent(event) {
			Present()
		}
	case sdl.EVENT_KEY_DOWN:
		kev := event.KeyboardEvent()
		if kev.Repeat {
			return
		}
		if kev.Mod&sdl.KMOD_ALT != 0 && kev.Scancode == sdl.SCANCODE_RETURN {
			toggleFullscreen()
			return
		}
		if isModifierKey(kev.Scancode) {
			return
		}
		keyPressed = true
		keyValue = normalizeKeyName(kev.Key.KeyName())
		keyLatch = keyValue
	case sdl.EVENT_KEY_UP:
		kev := event.KeyboardEvent()
		if keyValue == normalizeKeyName(kev.Key.KeyName()) {
			keyValue = ""
		}
	}
}

func renderChar(row, col int, c byte) {
	if c > 64 {
		c = 64 + (c & 0x7f)
	}
	src := sdl.FRect{X: float32((c & 0x0f) << 3), Y: float32((c >> 4) << 3), W: 8, H: 8}
	dst := sdl.FRect{X: float32(col << 3), Y: float32(row << 3), W: 8, H: 8}
	_ = renderer.RenderTexture(fontAtlas, &src, &dst)
}

func buildFontAtlas() {
	var pixels [192]byte
	for i := byte(0); i < 128; i++ {
		dst := sdl.Rect{X: int32((i & 0x0f) << 3), Y: int32((i >> 4) << 3), W: 8, H: 8}
		buildCharPixels(font[i], &pixels)
		_ = fontAtlas.Update(&dst, pixels[:], 8*3)
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
