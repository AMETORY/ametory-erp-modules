package utils

import (
	"bytes"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

// GeneratePDF generates a PDF file from an HTML string.
//
// Parameters:
//   - wkhtmltopdfPath: The file path for the wkhtmltopdf binary. If not empty, it sets the path to the wkhtmltopdf executable.
//   - footer: The footer text to be added to each page of the PDF. If empty, no footer is added.
//   - html: The HTML content to be converted into a PDF.
//
// Returns:
//   - A byte slice containing the generated PDF.
//   - An error if there is an issue during PDF generation.
//
// This function sets various options for the PDF, such as enabling local file access, setting page margins to zero, and
// disabling smart shrinking. It also configures footer settings if a footer is provided. It supports customization of
// the PDF size, DPI, and orientation.
func GeneratePDF(wkhtmltopdfPath, footer, html string) ([]byte, error) {
	if wkhtmltopdfPath != "" {
		wkhtmltopdf.SetPath(wkhtmltopdfPath)
	}
	// Create a new PDF generator
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	// Add HTML string to the PDF generator
	page := wkhtmltopdf.NewPageReader(bytes.NewReader([]byte(html)))

	if footer != "" {
		page.FooterLeft.Set(footer)
		page.FooterRight.Set("[page]/[topage]")
	}
	page.FooterFontSize.Set(8)
	page.FooterSpacing.Set(2)
	page.FooterFontName.Set("Ubuntu")
	// Set options for the page
	page.EnableLocalFileAccess.Set(true) // Enable local file access
	page.NoBackground.Set(false)         // Ensure background is not disabled
	page.DisableSmartShrinking.Set(true) // Disable smart shrinking
	// page.UserStyleSheet.Set("path/to/your/styles.css") // Optionally add custom stylesheet
	// Add the page to the PDF generator
	pdfg.AddPage(page)
	pdfg.MarginBottom.Set(0)
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.MarginTop.Set(0)

	// Set some options for the PDF
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfg.NoPdfCompression.Set(false)

	// Create PDF
	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Bytes(), nil
}
