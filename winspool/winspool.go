package winspool

import (
	"fmt"
	"goprintany/model"
	"image"
	"image/draw"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/transform"
)

var pageOrientationByType = map[model.PageOrientationType]int16{
	model.PageOrientationPortrait:  DMORIENT_PORTRAIT,
	model.PageOrientationLandscape: DMORIENT_LANDSCAPE,
}

func RawDataToPrinter(printerName, docName string, data []byte) (int32, error) {
	var count uint32 = 0
	var jobID int32 = 0
	// Open a handle to the printer.
	h, err := OpenPrinter(printerName)
	if err == nil {
		// Inform the spooler the document is beginning.
		jobID, err = h.StartDocPrinter(1, docName, "RAW")
		if err == nil {
			// Start a page.
			err = h.StartPagePrinter()
			if err == nil {
				// Send the data to the printer.
				count, err = h.WritePrinter(data)
				if err == nil {
					_ = h.EndPagePrinter()
				}
			}
			// Inform the spooler that the document is ending.
			_ = h.EndDocPrinter()
		}
		// Close the printer handle.
		err = h.ClosePrinter()
	}
	// Check to see if correct number of bytes were written.
	if count != uint32(len(data)) {
		err = fmt.Errorf("wrote %d bytes to the printer expected %d", count, len(data))
	}
	return jobID, err
}

func ImageDataToPrinter(printerName, docName string, imageData image.Image, imageSetting ImageSetting, printerSetting PrinterSetting) (int32, error) {
	hPrinter, err := OpenPrinter(printerName)
	if err != nil {
		return 0, err
	}
	devMode, err := hPrinter.DocumentPropertiesGet(printerName)
	if err != nil {
		hPrinter.ClosePrinter()
		return 0, err
	}

	printerSetting.Apply(devMode)

	if err = hPrinter.DocumentPropertiesSet(printerName, devMode); err != nil {
		hPrinter.ClosePrinter()
		return 0, err
	}

	if err = hPrinter.ClosePrinter(); err != nil {
		return 0, err
	}

	hDC, err := CreateDC(printerName, devMode)
	if err != nil {
		return 0, err
	}
	jobID, err := hDC.StartDoc(docName)
	if err != nil {
		hDC.DeleteDC()
		return 0, err
	}

	// Set device to zero offset, and to points scale.
	xDPI := hDC.GetDeviceCaps(LOGPIXELSX)
	yDPI := hDC.GetDeviceCaps(LOGPIXELSY)
	xMarginPixels := hDC.GetDeviceCaps(PHYSICALOFFSETX)
	yMarginPixels := hDC.GetDeviceCaps(PHYSICALOFFSETY)
	xform := NewXFORM(float32(xDPI)/float32(imageSetting.PPI), float32(yDPI)/float32(imageSetting.PPI), float32(-xMarginPixels), float32(-yMarginPixels))

	if err := hDC.SetGraphicsMode(GM_ADVANCED); err != nil {
		hDC.DeleteDC()
		return 0, err
	}
	if err := hDC.SetWorldTransform(xform); err != nil {
		hDC.DeleteDC()
		return 0, err
	}

	if err := hDC.StartPage(); err != nil {
		hDC.DeleteDC()
		return 0, err
	}

	if imageSetting.Rotation != 0 {
		// Rotate the image through Image API, but maybe we should cnsider to change printer orientation
		imageData = transform.Rotate(imageData, float64(imageSetting.Rotation),
			&transform.RotationOptions{ResizeBounds: true, Pivot: &image.Point{0, 0}})
	}

	if imageSetting.FitToPaper {
		bounds := imageData.Bounds()
		wPaperPixels := hDC.GetDeviceCaps(PHYSICALWIDTH)
		hPaperPixels := hDC.GetDeviceCaps(PHYSICALHEIGHT)
		wPrintablePixels := hDC.GetDeviceCaps(HORZRES)
		hPrintablePixels := hDC.GetDeviceCaps(VERTRES)
		w := bounds.Dx()
		h := bounds.Dy()

		scale := getScale(int32(w), int32(h), wPrintablePixels, hPrintablePixels, wPaperPixels, hPaperPixels, xDPI, yDPI, imageSetting.PPI)
		newW := int(float64(w) * scale)
		newH := int(float64(h) * scale)

		imageData = transform.Resize(imageData, newW, newH, transform.Gaussian)
	}

	bounds := imageData.Bounds()

	hbitmapImage := NewHbitMapImage(bounds.Dx(), bounds.Dy())
	// copy the image data into the hbitmapImage
	draw.Draw(hbitmapImage, hbitmapImage.Bounds(), imageData, image.Point{}, draw.Src)

	// Display the Image
	hdcMem, err := hDC.CreateCompatibleDC()
	if err != nil {
		return 0, err
	}
	hdcMem.SelectObject(HGDIOBJ(hbitmapImage.hbitmap))
	hDC.BitBlt(0, 0, int32(hbitmapImage.Rect.Dx()), int32(hbitmapImage.Rect.Dy()), hdcMem, 0, 0, SRCCOPY)

	hdcMem.DeleteDC()
	hDC.EndPage()
	hDC.EndDoc()
	hDC.DeleteDC()

	return jobID, err
}

func getManModel(driverName string) (man string, model string) {
	man = ""
	model = ""

	parts := strings.SplitN(driverName, " ", 2)
	if len(parts) > 0 && len(parts[0]) > 0 {
		man = parts[0]
	}
	if len(parts) > 1 && len(parts[1]) > 0 {
		model = parts[1]
	}

	return
}

func convertPrinterState(wsStatus uint32, wsAttributes uint32) *model.PrinterStateSection {
	state := model.PrinterStateSection{
		State:       model.CloudDeviceStateIdle,
		VendorState: &model.VendorState{},
	}

	if wsStatus&(PRINTER_STATUS_PRINTING|PRINTER_STATUS_PROCESSING) != 0 {
		state.State = model.CloudDeviceStateProcessing
	}

	if wsStatus&PRINTER_STATUS_PAUSED != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateWarning,
			DescriptionLocalized: model.NewLocalizedString("printer paused"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_ERROR != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("printer error"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_PENDING_DELETION != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("printer is being deleted"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_PAPER_JAM != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("paper jam"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_PAPER_OUT != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("paper out"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_MANUAL_FEED != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("manual feed mode"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_PAPER_PROBLEM != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("paper problem"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}

	// If PRINTER_ATTRIBUTE_WORK_OFFLINE is set
	// spooler won't despool any jobs to the printer.
	// At least for some USB printers, this flag is controlled
	// automatically by the system depending on the state of physical connection.
	if wsStatus&PRINTER_STATUS_OFFLINE != 0 || wsAttributes&PRINTER_ATTRIBUTE_WORK_OFFLINE != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("printer is offline"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_IO_ACTIVE != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("active I/O state"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_BUSY != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("busy"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_OUTPUT_BIN_FULL != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("output bin is full"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_NOT_AVAILABLE != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("printer not available"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_WAITING != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("waiting"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_INITIALIZING != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("intitializing"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_WARMING_UP != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("warming up"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_TONER_LOW != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateWarning,
			DescriptionLocalized: model.NewLocalizedString("toner low"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_NO_TONER != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("no toner"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_PAGE_PUNT != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("cannot print the current page"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_USER_INTERVENTION != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("user intervention required"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_OUT_OF_MEMORY != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("out of memory"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_DOOR_OPEN != 0 {
		state.State = model.CloudDeviceStateStopped
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("door open"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_SERVER_UNKNOWN != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateError,
			DescriptionLocalized: model.NewLocalizedString("printer status unknown"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}
	if wsStatus&PRINTER_STATUS_POWER_SAVE != 0 {
		vs := model.VendorStateItem{
			State:                model.VendorStateInfo,
			DescriptionLocalized: model.NewLocalizedString("power save mode"),
		}
		state.VendorState.Item = append(state.VendorState.Item, vs)
	}

	if len(state.VendorState.Item) == 0 {
		state.VendorState = nil
	}

	return &state
}

func convertMediaSize(printerName, portName string, devMode *DevMode) (*model.MediaSize, error) {
	defSize, defSizeOK := devMode.GetPaperSize()
	defLength, defLengthOK := devMode.GetPaperLength()
	defWidth, defWidthOK := devMode.GetPaperWidth()

	names, err := DeviceCapabilitiesStrings(printerName, portName, DC_PAPERNAMES, 64*2)
	if err != nil {
		return nil, err
	}
	papers, err := DeviceCapabilitiesUint16Array(printerName, portName, DC_PAPERS)
	if err != nil {
		return nil, err
	}
	sizes, err := DeviceCapabilitiesInt32Pairs(printerName, portName, DC_PAPERSIZE)
	if err != nil {
		return nil, err
	}
	if len(names) != len(papers) || len(names) != len(sizes)/2 {
		return nil, nil
	}

	ms := model.MediaSize{
		Option: make([]model.MediaSizeOption, 0, len(names)),
	}

	var foundDef bool
	for i := range names {
		if names[i] == "" {
			continue
		}
		// Convert from tenths-of-mm to micrometers
		width, length := sizes[2*i]*100, sizes[2*i+1]*100

		var def bool
		if !foundDef {
			if defSizeOK {
				if uint16(defSize) == papers[i] {
					def = true
					foundDef = true
				}
			} else if defLengthOK && int32(defLength) == length && defWidthOK && int32(defWidth) == width {
				def = true
				foundDef = true
			}
		}

		o := model.MediaSizeOption{
			Name:                       model.MediaSizeCustom,
			WidthMicrons:               width,
			HeightMicrons:              length,
			IsDefault:                  def,
			VendorID:                   strconv.FormatUint(uint64(papers[i]), 10),
			CustomDisplayNameLocalized: model.NewLocalizedString(names[i]),
		}
		ms.Option = append(ms.Option, o)
	}

	if !foundDef && len(ms.Option) > 0 {
		ms.Option[0].IsDefault = true
	}

	return &ms, nil
}

func getScale(wDocPoints, hDocPoints, wPrintablePixels, hPrintablePixels, wPaperPixels, hPaperPixels, xDPI, yDPI, ppi int32) (scale float64) {

	wPrintablePoints, hPrintablePoints := float64(wPrintablePixels*ppi)/float64(xDPI), float64(hPrintablePixels*ppi)/float64(yDPI)

	xScale, yScale := wPrintablePoints/float64(wDocPoints), hPrintablePoints/float64(hDocPoints)
	if xScale < yScale {
		scale = xScale
	} else {
		scale = yScale
	}
	return
}

// GetPrinters gets all Windows printers found on this computer.
func GetPrinters() ([]model.Printer, error) {
	pi2s, err := EnumPrinters2()
	if err != nil {
		return nil, err
	}

	printers := make([]model.Printer, 0, len(pi2s))
	for _, pi2 := range pi2s {
		printerName := pi2.GetPrinterName()
		portName := pi2.GetPortName()
		devMode := pi2.GetDevMode()
		location := pi2.GetLocation()

		manufacturer, model1 := getManModel(pi2.GetDriverName())
		printer := model.Printer{
			Name:               printerName,
			DefaultDisplayName: printerName,
			Manufacturer:       manufacturer,
			Model:              model1,
			State:              convertPrinterState(pi2.GetStatus(), pi2.GetAttributes()),
			Description:        &model.PrinterDescriptionSection{},
			Location:           location,
			// Tags: map[string]string{
			// 	"printer-location": pi2.GetLocation(),
			// },
		}

		// Advertise color based on default value, which should be a solid indicator
		// of color-ness, because the source of this devMode object is EnumPrinters.
		if def, ok := devMode.GetColor(); ok {
			if def == DMCOLOR_COLOR {
				printer.Description.Color = &model.Color{
					Option: []model.ColorOption{
						{
							VendorID:                   strconv.FormatInt(int64(DMCOLOR_COLOR), 10),
							Type:                       model.ColorTypeStandardColor,
							IsDefault:                  true,
							CustomDisplayNameLocalized: model.NewLocalizedString("Color"),
						},
						{
							VendorID:                   strconv.FormatInt(int64(DMCOLOR_MONOCHROME), 10),
							Type:                       model.ColorTypeStandardMonochrome,
							IsDefault:                  false,
							CustomDisplayNameLocalized: model.NewLocalizedString("Monochrome"),
						},
					},
				}
			} else if def == DMCOLOR_MONOCHROME {
				printer.Description.Color = &model.Color{
					Option: []model.ColorOption{
						{
							VendorID:                   strconv.FormatInt(int64(DMCOLOR_MONOCHROME), 10),
							Type:                       model.ColorTypeStandardMonochrome,
							IsDefault:                  true,
							CustomDisplayNameLocalized: model.NewLocalizedString("Monochrome"),
						},
					},
				}
			}
		}

		if def, ok := devMode.GetDuplex(); ok {
			duplex, err := DeviceCapabilitiesInt32(printerName, portName, DC_DUPLEX)
			if err != nil {
				return nil, err
			}
			if duplex == 1 {
				printer.Description.Duplex = &model.Duplex{
					Option: []model.DuplexOption{
						{
							Type:      model.DuplexNoDuplex,
							IsDefault: def == DMDUP_SIMPLEX,
						},
						{
							Type:      model.DuplexLongEdge,
							IsDefault: def == DMDUP_VERTICAL,
						},
						{
							Type:      model.DuplexShortEdge,
							IsDefault: def == DMDUP_HORIZONTAL,
						},
					},
				}
			}
		}

		if def, ok := devMode.GetOrientation(); ok {
			orientation, err := DeviceCapabilitiesInt32(printerName, portName, DC_ORIENTATION)
			if err != nil {
				return nil, err
			}
			if orientation == 90 || orientation == 270 {
				printer.Description.PageOrientation = &model.PageOrientation{
					Option: []model.PageOrientationOption{
						{
							Type:      model.PageOrientationPortrait,
							IsDefault: def == DMORIENT_PORTRAIT,
						},
						{
							Type:      model.PageOrientationLandscape,
							IsDefault: def == DMORIENT_LANDSCAPE,
						},
					},
				}
			}
		}

		if def, ok := devMode.GetCopies(); ok {
			copies, err := DeviceCapabilitiesInt32(printerName, portName, DC_COPIES)
			if err != nil {
				return nil, err
			}
			if copies > 1 {
				printer.Description.Copies = &model.Copies{
					Default: int32(def),
					Max:     copies,
				}
			}
		}

		printer.Description.MediaSize, err = convertMediaSize(printerName, portName, devMode)
		if err != nil {
			return nil, err
		}

		if def, ok := devMode.GetCollate(); ok {
			collate, err := DeviceCapabilitiesInt32(printerName, portName, DC_COLLATE)
			if err != nil {
				return nil, err
			}
			if collate == 1 {
				printer.Description.Collate = &model.Collate{
					Default: def == DMCOLLATE_TRUE,
				}
			}
		}

		printers = append(printers, printer)
	}

	return printers, nil
}

type PrinterSetting struct {
	PageOrientation model.PageOrientationType
	Copy            int16
}

func (p PrinterSetting) Apply(devMode *DevMode) {
	if pageOrientation, ok := pageOrientationByType[p.PageOrientation]; ok {
		devMode.SetOrientation(pageOrientation)
	}
	if p.Copy > 0 {
		devMode.SetCopies(int16(p.Copy))
	}
}

type ImageSetting struct {
	PPI        int32
	Rotation   int32
	Scale      float32
	FitToPaper bool
}
