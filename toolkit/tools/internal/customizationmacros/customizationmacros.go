// Copyright Microsoft Corporation.
// Licensed under the MIT License.

/*
customizationmacros package provides functions to add customization macros to a root directory. These macros are saved in the
root directory's default rpm macros directory. Each customization has a corresponding macro file that is generated as needed.
*/
package customizationmacros

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/file"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/logger"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/rpm"
)

const (
	// macro files to customize the installer and final images
	disableRpmDocsMacroFile      = "macros.installercustomizations_disable_docs"
	configureRpmLocalesMacroFile = "macros.installercustomizations_customize_locales"
)

var (
	// Universal header comments for all customization macro files
	customizationMacroHeaderComments = []string{
		"This macro file was dynamically generated by the Azure Linux Toolkit image generator",
		"based on the configuration used at image creation time.",
		"",
	}
	docComments []string = []string{
		"This stops anything rpm considers a documentation files from being installed.",
		"To enable documentation files, remove this file, or comment out '%%_excludedocs 1'",
		"Any packages which are already installed must be reinstalled for this change to take effect.",
	}
	localeComments []string = []string{
		"This stops locale files from being installed. %%_install_langs acts as a filter for locales",
		"which start with the provides strings. Setting it to an invalid value (ie 'NONE') will",
		"prevent any locale files from being installed.",
		"To enable locale files, remove this file, or comment out '%%_install_langs <LOCALE STRING>'",
		"Any packages which are already installed must be reinstalled for this change to take effect.",
	}
)

// AddCustomizationMacros adds the currently defined image custimization macros to the specified root directory.
// For each of disableRpmDocs and overrideRpmLocales a macro file is created with the corresponding macros defined in the
// default rpm macros directory.
func AddCustomizationMacros(rootDir string, disableRpmDocs bool, overrideRpmLocales string) (err error) {
	macroDir, err := rpm.GetMacroDir()
	if err != nil {
		return fmt.Errorf("failed to get rpm macro directory when adding customization macros:\n%w", err)
	}
	fullMacroDirPath := filepath.Join(rootDir, macroDir)

	if disableRpmDocs {
		logger.Log.Debugf("Disabling documentation packages")
		err = AddMacroFile(fullMacroDirPath, rpm.DisableDocumentationDefines(), disableRpmDocsMacroFile, docComments)
		if err != nil {
			return fmt.Errorf("failed to add disable docs macro file:\n%w", err)
		}
	}
	if overrideRpmLocales != "" {
		logger.Log.Debugf("Overriding locale packages with (%s)", overrideRpmLocales)
		err = AddMacroFile(fullMacroDirPath, rpm.OverrideLocaleDefines(overrideRpmLocales), configureRpmLocalesMacroFile, localeComments)
		if err != nil {
			return fmt.Errorf("failed to add override locales macro file:\n%w", err)
		}
	}
	return nil
}

// formatComments ensures that the comments are valid for a macro file: ie they are empty or start with '#'
func formatComments(comments []string) (formattedComments []string) {
	for _, comment := range comments {
		if strings.TrimSpace(comment) == "" {
			formattedComments = append(formattedComments, "")
		} else {
			formattedComments = append(formattedComments, strings.TrimSpace("# "+comment))
		}
	}
	return formattedComments
}

// AddMacroFile adds a macro file to the specified root directory with the specified macros. The macro file
// is created in the default rpm macros directory. The macro file will include a default header, with additional comments
// if desired. Each extra comment should start with a '#' character.
func AddMacroFile(macroDir string, macros map[string]string, macroFileName string, extraComments []string) error {
	if len(macros) == 0 {
		return nil
	}

	header := customizationMacroHeaderComments
	if len(extraComments) > 0 {
		extraComments = append(extraComments, "")
		header = append(customizationMacroHeaderComments, extraComments...)
	}

	header = formatComments(header)

	macroFilePath := filepath.Join(macroDir, macroFileName)
	err := os.MkdirAll(macroDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory for macro file:\n%w", err)
	}

	macroLines := []string{}
	for key, value := range macros {
		macroLines = append(macroLines, fmt.Sprintf("%%%s %s", key, value))
	}
	// Sort the lines to ensure the macro file is deterministic
	sort.Strings(macroLines)

	// Add the header, followed by any additional comments to the top of the file
	finalLines := append(header, macroLines...)
	err = file.WriteLines(finalLines, macroFilePath)
	if err != nil {
		return fmt.Errorf("failed to write macro file:\n%w", err)
	}
	return nil
}
