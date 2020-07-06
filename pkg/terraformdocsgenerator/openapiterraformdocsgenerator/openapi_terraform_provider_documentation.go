package openapiterraformdocsgenerator

import (
	"io"
)

// TerraformProviderDocumentation defines the attributes needed to generate Terraform provider documentation
type TerraformProviderDocumentation struct {
	ProviderName                string
	ProviderNotes               []string
	ProviderInstallation        ProviderInstallation
	ProviderConfiguration       ProviderConfiguration
	ProviderResources           ProviderResources
	DataSources                 DataSources
	ShowSpecialTermsDefinitions bool
}

// RenderHTML writes to the given writer argument the Terraform provider documentation
func (t TerraformProviderDocumentation) RenderHTML(w io.Writer) error {
	return t.renderZendeskHTML(w, TableOfContentsTmpl, ProviderInstallationTmpl, ProviderConfigurationTmpl, ProviderResourcesTmpl, DataSourcesTmpl, SpecialTermsTmpl)
}

// RenderZendeskHTML renders the documentation in HTML
func (t TerraformProviderDocumentation) renderZendeskHTML(w io.Writer, tableOfContentsTemplate, providerInstallationTemplate, providerConfigurationTemplate, providerResourcesConfiguration, providerDatSourcesTemplate, specialTermsDefinitionsTemplate string) error {
	err := render(w, "TerraformProviderDocTableOfContents", tableOfContentsTemplate, t)
	if err != nil {
		return err
	}
	err = render(w, "ProviderInstallation", providerInstallationTemplate, t.ProviderInstallation)
	if err != nil {
		return err
	}
	err = render(w, "ProviderConfiguration", providerConfigurationTemplate, t.ProviderConfiguration)
	if err != nil {
		return err
	}
	err = render(w, "ProviderResources", providerResourcesConfiguration, t.ProviderResources)
	if err != nil {
		return err
	}
	err = render(w, "DataSources", providerDatSourcesTemplate, t.DataSources)
	if err != nil {
		return err
	}
	err = render(w, "TerraformProviderDocSpecialTermsDefinitions", specialTermsDefinitionsTemplate, t)
	if err != nil {
		return err
	}
	return nil
}
