// package frontmatter includes types and functions that provide resource plan
// modifiers for using Markdown front matter for attribute values. Front matter
// is an alternative way to specify parameters for changelogs, custom pages,
// and docs.
package frontmatter

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/liveoaklabs/readme-api-go-client/readme"
)

// ReadmeFrontMatter represents the front matter keys available to ReadMe changelogs, custom pages, and docs.
type ReadmeFrontMatter struct {
	Body          string                `yaml:"body,omitempty"`          // changelogs, custom pages, docs
	Category      string                `yaml:"category"`                // docs
	CategorySlug  string                `yaml:"categorySlug"`            // docs
	Error         readme.DocErrorObject `yaml:"error,omitempty"`         // docs
	Hidden        *bool                 `yaml:"hidden"`                  // changelogs, custom pages, docs
	HTML          string                `yaml:"html,omitempty"`          // custom page
	HTMLMode      *bool                 `yaml:"htmlmode"`                // custom page
	Order         int64                 `yaml:"order"`                   // docs
	ParentDoc     string                `yaml:"parentDoc,omitempty"`     // docs
	ParentDocSlug string                `yaml:"parentDocSlug,omitempty"` // docs
	Title         string                `yaml:"title"`                   // changelogs, custom pages, docs
	Type          string                `yaml:"type,omitempty"`          // changelogs, docs
}

// GetValue parses the 'body' attribute value for Markdown front matter and
// returns a specified key's value if it's present in the front matter.
//
// The `attribute` parameter is the struct field name representing the YAML
// front matter key.
//
// This returns a `reflect.Value` to be evaluated as needed based on the type
// and other conditions.
//
// A string value is provided in place of an error for use with the plugin
// framework's diagnostics package.
func GetValue(ctx context.Context, body, attribute string) (reflect.Value, string) {
	tflog.Debug(ctx, fmt.Sprintf("checking body front matter for attribute '%s'", attribute))

	// Get the FrontMatter from the "body" attribute.
	frontMatter := ReadmeFrontMatter{}
	_, err := frontmatter.Parse(strings.NewReader(body), &frontMatter)
	if err != nil {
		return reflect.Value{}, err.Error()
	}

	tflog.Debug(ctx, fmt.Sprintf("body front matter=%+v", frontMatter))

	// Get the field value matching the attribute.
	v := reflect.ValueOf(frontMatter)
	field := v.FieldByName(attribute)

	// If the field does not exist, return an empty value.
	if field == (reflect.Value{}) {
		return reflect.Value{}, ""
	}

	// If the field exists and is empty, return an empty value.
	if field.IsZero() {
		tflog.Debug(ctx, fmt.Sprintf("no front matter found for attribute %s", attribute))

		return reflect.Value{}, ""
	}

	tflog.Info(ctx, fmt.Sprintf("found front matter for attribute %s", attribute))

	return field, ""
}
