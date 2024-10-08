// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

//

package imagecustomizerapi

import (
	"fmt"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/sliceutils"
	"gopkg.in/yaml.v3"
)

// DestinationFileConfigList is a list of destination files where the source file will be copied to in the final image.
// This type exists to allow a custom marshaller to be attached to it.
type FileConfigList []FileConfig

// FileConfig specifies options for how a file is copied in the target OS.
type FileConfig struct {
	// The file path in the target OS that the file will be copied to.
	Path string `yaml:"path"`

	// The file permissions to set on the file.
	Permissions *FilePermissions `yaml:"permissions"`
}

var (
	fileConfigValidFields = []string{"path", "permissions"}
)

func (l *FileConfigList) IsValid() (err error) {
	if len(*l) <= 0 {
		return fmt.Errorf("list is empty")
	}

	for i, fileConfig := range *l {
		err = fileConfig.IsValid()
		if err != nil {
			return fmt.Errorf("invalid fileConfig at index %d:\n%w", i, err)
		}
	}

	return nil
}

func (l *FileConfigList) UnmarshalYAML(value *yaml.Node) error {
	var err error

	// Try to parse as a single value.
	var fileConfig FileConfig
	err = value.Decode(&fileConfig)
	if err == nil {
		*l = FileConfigList{fileConfig}
		return nil
	}

	// Try to parse as a list.
	type IntermediateTypeFileConfigList FileConfigList
	err = value.Decode((*IntermediateTypeFileConfigList)(l))
	if err != nil {
		return fmt.Errorf("failed to parse fileConfigList:\n%w", err)
	}

	return nil
}

func (f *FileConfig) IsValid() (err error) {
	// Path
	if f.Path == "" {
		return fmt.Errorf("invalid path value: empty string")
	}

	// Permissions
	if f.Permissions != nil {
		err = f.Permissions.IsValid()
		if err != nil {
			return fmt.Errorf("invalid permissions value:\n%w", err)
		}
	}

	return nil
}

func (f *FileConfig) UnmarshalYAML(value *yaml.Node) error {
	var err error

	if value.Kind == yaml.ScalarNode {
		// Parse as a string.
		*f = FileConfig{
			Path:        value.Value,
			Permissions: nil,
		}
		return nil
	}

	// Parse as a struct.
	*f = FileConfig{}

	type IntermediateTypeFileConfig FileConfig
	err = value.Decode((*IntermediateTypeFileConfig)(f))
	if err != nil {
		return fmt.Errorf("failed to parse fileConfig:\n%w", err)
	}

	// yaml.Node.Decode() doesn't respect the KnownFields() option.
	// So, manually enforce this.
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		if !sliceutils.ContainsValue(fileConfigValidFields, key) {
			return fmt.Errorf("line %d: field %s not found in type %s", value.Line, key, "FileConfig")
		}
	}

	return nil
}
