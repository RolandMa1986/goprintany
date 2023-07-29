package winspool

import (
	"image"
	"unsafe"

	"github.com/lxn/win"
)

type HbitMapImage struct {
	BGRA
	hbitmap win.HBITMAP
}

func NewHbitMapImage(width, height int) *HbitMapImage {
	this := &HbitMapImage{}
	this.Rect = image.Rect(0, 0, width, height)
	this.Stride = 4 * width // 4 bytes per pixel
	this.Pix = make([]uint8, this.Stride*height)
	this.hbitmap, this.Pix = CreateHBitmap(width, height)

	return this
}

func CreateHBitmap(width, height int) (win.HBITMAP, []uint8) {
	var bi win.BITMAPV5HEADER
	const bitsPerPixel = 32 // We only work with 32 bit (BGRA)
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(width)
	bi.BiHeight = -int32(height)
	bi.BiPlanes = 1
	bi.BiBitCount = uint16(bitsPerPixel)
	bi.BiCompression = win.BI_RGB
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	tempdc := win.GetDC(0)
	defer win.ReleaseDC(0, tempdc)

	var lpBits unsafe.Pointer

	// Create the DIB section with an alpha channel.
	hbitmap := win.CreateDIBSection(tempdc, &bi.BITMAPINFOHEADER, win.DIB_RGB_COLORS, &lpBits, 0, 0)
	switch hbitmap {
	case 0, win.ERROR_INVALID_PARAMETER:
		return 0, nil
	}

	length := ((width*bitsPerPixel + 31) / 32) * 4 * height
	// Slice memory layout
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(lpBits), length, length}
	// Use unsafe to turn sl into a []uint8
	bitmapArray := *(*[]uint8)(unsafe.Pointer(&sl))
	return hbitmap, bitmapArray
}
