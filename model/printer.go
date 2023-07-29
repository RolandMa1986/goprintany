package model

// DuplexVendorMap maps a DuplexType to a CUPS key:value option string for a given printer.
type DuplexVendorMap map[DuplexType]string

type Printer struct {
	Name               string                     // CUPS: cups_dest_t.name (CUPS key);
	DefaultDisplayName string                     // CUPS: printer-info;
	Manufacturer       string                     // CUPS: PPD;
	Model              string                     // CUPS: PPD;
	Location           string                     // CUPS: printer-location;
	State              *PrinterStateSection       // CUPS: various;
	Description        *PrinterDescriptionSection // CUPS: translated PPD;
	CapsHash           string                     // CUPS: hash of PPD;
	Tags               map[string]string          // CUPS: all printer attributes;
	DuplexMap          DuplexVendorMap            // CUPS: PPD;
	QuotaEnabled       bool
	DailyQuota         int
}
