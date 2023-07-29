## Go Print Any

Go Print Any is a simple cli tool writen in Go to print images or other file on windows.

## Why？

On linux, you can use `lp` to print images and there are many options. Although you can use `mspaint` or `ImageView_PrintTo ` to print images on windows, but you have no control on how to print the image. 

## How to use?

```bash
$ goprintany.exe
print a file

Usage:
  print print [flags]

Flags:
  -r, --Rotation int32   Rotation angle in degrees,eg: 0, 90, 180, 270
  -s, --Scale float32    Scale factor of the image,don't use with fitToPaper (default 1)
  -f, --file string      file path
  -a, --fitToPaper       fit to paper
  -h, --help             help for print
  -i, --ppi int32        Pixels per inch of the image (default 96)
  -p, --printer string   Printer Name
  -t, --type string      Data type send to printer：RAW，IMAGE，PDF (default "IMAGE")
```

## reference

This project is inspired by:

[Google Cloud Print Connector](https://github.com/google/cloud-print-connector)

[GoPaint](https://github.com/shahfarhadreza/gopaint)
