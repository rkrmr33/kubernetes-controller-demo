/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	DefaultCatDuration = DurationString((time.Second * 5).String())

	DefaultCatMessage = "hello, world!"

	DefaultCatLives int32 = 9

	log = logf.Log.WithName("cat-resource")
)

func (r *Cat) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-example-cats-io-example-cats-io-v1alpha1-cat,mutating=true,failurePolicy=fail,sideEffects=None,groups=example.cats.io,resources=cats,verbs=create;update,versions=v1alpha1,name=mcat.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Cat{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *Cat) Default() {
	log.Info("default", "name", c.Name)

	if c.Spec.Duration == nil {
		c.Spec.Duration = &DefaultCatDuration
	}

	if c.Spec.Message == nil {
		c.Spec.Message = &DefaultCatMessage
	}

	if c.Spec.TotalLives == nil {
		c.Spec.TotalLives = &DefaultCatLives
	}
}

//+kubebuilder:webhook:path=/validate-example-cats-io-example-cats-io-v1alpha1-cat,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.cats.io,resources=cats,verbs=create;update,versions=v1alpha1,name=vcat.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cat{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *Cat) ValidateCreate() error {
	log.Info("validate create", "name", c.Name)
	validateCat(c)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *Cat) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", c.Name)
	validateCat(c)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cat) ValidateDelete() error {
	log.Info("validate delete", "name", r.Name)

	return nil
}

func validateCat(c *Cat) error {
	var allErrs field.ErrorList

	if err := validateCatDuration(string(*c.Spec.Duration), field.NewPath("spec").Child("duration")); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: SchemeBuilder.GroupVersion.Group, Kind: "Cat"}, c.Name, allErrs)
}

func validateCatDuration(dur string, fieldPath *field.Path) *field.Error {
	_, err := time.ParseDuration(dur)
	if err != nil {
		return field.Invalid(fieldPath, dur, "invalid duration value")
	}
	return nil
}
