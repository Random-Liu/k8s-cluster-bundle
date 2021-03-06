// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validate

import (
	"fmt"
	"regexp"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/blang/semver"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	apiVersionPattern = regexp.MustCompile(`^bundle.gke.io/\w+$`)
)

// Components validates a list of components.
func Components(components []*bundle.Component) field.ErrorList {
	errs := field.ErrorList{}
	for _, component := range components {
		compErrs := Component(component)
		for _, err := range compErrs {
			errs = append(errs, err)
		}
	}
	return errs
}

// Component validates a single component.
func Component(c *bundle.Component) field.ErrorList {
	errs := field.ErrorList{}
	pi := field.NewPath("Component")

	n := c.Spec.ComponentName
	if n == "" {
		errs = append(errs, field.Required(pi.Child("Spec", "ComponentName"), "components must have ComponentName"))
	}
	ver := c.Spec.Version
	if ver == "" {
		errs = append(errs, field.Required(pi.Child("Spec", "Version"), "components must have a Version"))
	}
	if n == "" || ver == "" {
		// Subsequent validation relies on components having a unique name+version pair.
		return errs
	}

	p := cPath(c.ComponentReference())

	if nameErrs := validateName(p.Child("Spec", "ComponentName"), n); len(errs) > 0 {
		errs = append(errs, nameErrs...)
	}

	api := c.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, field.Invalid(p.Child("APIVersion"), api, "must have the form \"bundle.gke.io/<version>\""))
	}

	expType := "Component"

	if k := c.Kind; k != expType {
		errs = append(errs, field.Invalid(p.Child("Kind"), k, "kind must be Component"))
	}

	if _, err := semver.Parse(ver); err != nil {
		errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, fmt.Sprintf("must be a SemVer version: %v", err)))
	}

	return errs
}

// ComponentSet validates a component.
func ComponentSet(cs *bundle.ComponentSet) field.ErrorList {
	p := field.NewPath("ComponentSet")

	errs := field.ErrorList{}
	if cs == nil {
		return errs
	}

	n := cs.Spec.SetName
	if n == "" {
		errs = append(errs, field.Required(p.Child("Spec", "SetName"), "setName is required"))
	}
	ver := cs.Spec.Version
	if ver == "" {
		errs = append(errs, field.Required(p.Child("Spec", "Version"), "version is required"))
	}

	if n == "" || ver == "" {
		// Other validation relies on components having a unique name+version pair.
		return errs
	}

	if nameErrs := validateName(p.Child("Spec", "SetName"), n); len(nameErrs) > 0 {
		errs = append(errs, nameErrs...)
	}

	api := cs.APIVersion
	if !apiVersionPattern.MatchString(api) {
		errs = append(errs, field.Invalid(p.Child("APIVersion"), api, "must have an apiVersion of the form \"bundle.gke.io/<version>\""))
	}

	expType := "ComponentSet"
	if k := cs.Kind; k != expType {
		errs = append(errs, field.Invalid(p.Child("Kind"), k, "must be ComponentSet"))
	}

	if _, err := semver.Parse(ver); err != nil {
		errs = append(errs, field.Invalid(p.Child("Spec", "Version"), ver, fmt.Sprintf("must be a SemVer version: %v", err)))
	}

	return errs
}

func cPath(ref bundle.ComponentReference) *field.Path {
	return field.NewPath("Component").Key(fmt.Sprintf("%v", ref))
}
