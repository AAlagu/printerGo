package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"syscall"
	"unsafe"
)

var (
	dll               = syscall.MustLoadDLL("winspool.drv")
	getDefaultPrinter = dll.MustFindProc("GetDefaultPrinterW")
	openPrinter       = dll.MustFindProc("OpenPrinterW")
	closePrinter      = dll.MustFindProc("ClosePrinter")
	startPagePrinter  = dll.MustFindProc("StartPagePrinter")
	endPagePrinter    = dll.MustFindProc("EndPagePrinter")
	endDocPrinter     = dll.MustFindProc("EndDocPrinter")
	startDocPrinter   = dll.MustFindProc("StartDocPrinterW")
	writePrinter      = dll.MustFindProc("WritePrinter")
)

type DOC_INFO_1 struct {
	pDocName    *uint16
	pOutputFile *uint16
	pDatatype   *uint16
}

func main() {

	printerName, _ := GetDefaultPrinterName()
	//open the printer
	printerHandle, err := GoOpenPrinter(printerName)
	if err != nil {
		log.Fatalln("Failed to open printer")
	}
	defer GoClosePrinter(printerHandle)

	filePath := "D:/Lebenslauf.pdf"

	//Send to printer:
	err = GoPrint(printerHandle, filePath)
	if err != nil {
		log.Fatalln("during the func sendToPrinter, there was an error")
	}
}

func GetDefaultPrinterName() (string, []uint16) {
	var pn [256]uint16
	plen := len(pn)
	getDefaultPrinter.Call(uintptr(unsafe.Pointer(&pn)), uintptr(unsafe.Pointer(&plen)))
	printerName := syscall.UTF16ToString(pn[:])
	fmt.Println("Printer name:", printerName)
	printer16 := syscall.StringToUTF16(printerName)
	return printerName, printer16
}

func GoOpenPrinter(printerName string) (uintptr, error) {
	printerName16 := syscall.StringToUTF16(printerName)
	printerHandle, err := openPrinterFunc(printerName, printerName16)
	if err != nil {
		return 0, err
	}

	return printerHandle, nil
}

func openPrinterFunc(printerName string, printerName16 []uint16) (uintptr, error) {

	var printerHandle uintptr
	_, _, msg := openPrinter.Call(uintptr(unsafe.Pointer(&printerName16[0])), uintptr(unsafe.Pointer(&printerHandle)), 0)
	fmt.Println("open printer: ", msg)

	if printerHandle == 0 {
		return 0, fmt.Errorf("couldn't find printer: printerName")
	}

	return printerHandle, nil

}

func GoClosePrinter(printerHandle uintptr) {

	closePrinter.Call(printerHandle)

	//return
}

func GoPrint(printerHandle uintptr, path string) error {
	var err error

	startPrinter(printerHandle, path)
	startPagePrinter.Call(printerHandle)
	err = writePrinterFunc(printerHandle, path)
	endPagePrinter.Call(printerHandle)
	endDocPrinter.Call(printerHandle)

	return err
}

func startPrinter(printerHandle uintptr, path string) {

	/*arr := strings.Split(path, "/")
	l := len(arr)
	name := arr[l-1]*/

	d := DOC_INFO_1{
		pDocName:    &(syscall.StringToUTF16("Output pdf file"))[0],
		pOutputFile: nil,
		// Win7
		//pDataType = "RAW";
		// Win8+
		// pDataType = "XPS_PASS";
		pDatatype: &(syscall.StringToUTF16("XPS_PASS"))[0],
	}
	r1, r2, err := startDocPrinter.Call(printerHandle, 1, uintptr(unsafe.Pointer(&d)))
	fmt.Println("startDocPrinter: ", r1, r2, err)

	//return
}

func writePrinterFunc(printerHandle uintptr, path string) error {
	fmt.Println("About to write file to path: ", path)
	fileContents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var contentLen uintptr = uintptr(len(fileContents))
	fmt.Println("contentLen", contentLen)
	var writtenLen int
	_, _, err = writePrinter.Call(printerHandle, uintptr(unsafe.Pointer(&fileContents[0])), contentLen, uintptr(unsafe.Pointer(&writtenLen)))
	fmt.Println("Writing to printer:", err)

	return nil
}
