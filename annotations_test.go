// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/ogen-go/ogen"
	"github.com/stretchr/testify/assert"
)

func TestGetAnnotation(t *testing.T) {
	// True.
	assert.Equal(
		t,
		ptr(true),
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): WithPagination(true),
		}}).Pagination,
	)

	// False.
	assert.Equal(
		t,
		ptr(false),
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): WithPagination(false),
		}}).Pagination,
	)

	// Unspecified.
	var ptrBoolNil *bool
	assert.Equal(
		t,
		ptrBoolNil,
		GetAnnotation(&gen.Type{Annotations: map[string]any{
			Annotation{}.Name(): Annotation{},
		}}).Pagination,
	)

	// Test fields.
	assert.True(
		t,
		GetAnnotation(&gen.Field{Annotations: map[string]any{
			Annotation{}.Name(): WithSortable(true),
		}}).Sortable,
	)

	// Test edges.
	assert.Equal(
		t,
		ptr(true),
		GetAnnotation(&gen.Edge{Annotations: map[string]any{
			Annotation{}.Name(): WithEagerLoad(true),
		}}).EagerLoad,
	)
}

func TestValidateAnnotation(t *testing.T) {
	tests := []struct {
		name    string
		value   *gen.Type
		wantErr bool
	}{
		{
			name:  "no-annotation",
			value: &gen.Type{Annotations: map[string]any{}},
		},
		{
			name: "valid-annotation",
			value: &gen.Type{Annotations: map[string]any{
				Annotation{}.Name(): WithPagination(true), // Type's should support pagination.
			}},
			wantErr: false,
		},
		{
			name: "invalid-annotation-type-with-edge",
			value: &gen.Type{Annotations: map[string]any{
				Annotation{}.Name(): WithEagerLoad(false), // Only edges support eager loading.
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnnotations(tt.value)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAnnotation_Merge(t *testing.T) {
	tests := []struct {
		name        string
		annotations []Annotation
		want        Annotation
	}{
		{
			name: "no-annotations",
			annotations: []Annotation{
				{},
				{},
			},
			want: Annotation{},
		},
		{
			name: "overlap-single",
			annotations: []Annotation{
				WithPagination(true),
				WithPagination(false),
			},
			want: Annotation{
				Pagination: ptr(false),
			},
		},
		{
			name: "overlap-multiple",
			annotations: []Annotation{
				WithPagination(false),
				WithDescription("foo"),
				WithPagination(true),
				WithDescription("bar"),
			},
			want: Annotation{
				Pagination:  ptr(true),
				Description: "bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out Annotation
			for _, a := range tt.annotations {
				out, _ = out.Merge(a).(Annotation)
			}
			assert.Equal(t, tt.want, out)
		})
	}
}

func TestAnnotation_AdditionalTags(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet", WithAdditionalTags("Foo"))
			injectAnnotations(t, g, "Pet.categories", WithAdditionalTags("Bar"))
			return nil
		},
	})

	assert.Contains(t, r.json(`$.paths./pets.get.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets.post.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.get.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.patch.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}.delete.tags`), "Foo")
	assert.Contains(t, r.json(`$.paths./pets/{petID}/categories.get.tags`), "Bar")
}

func TestAnnotation_Tags(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet", WithTags("Foo"))
			injectAnnotations(t, g, "Pet.categories", WithTags("Bar"))
			return nil
		},
	})

	assert.Equal(t, "Foo", r.json(`$.paths./pets.get.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets.post.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.get.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.patch.tags.*`))
	assert.Equal(t, "Foo", r.json(`$.paths./pets/{petID}.delete.tags.*`))
	assert.Equal(t, "Bar", r.json(`$.paths./pets/{petID}/categories.get.tags.*`))
}

func TestAnnotation_EdgeUpdateBulk(t *testing.T) {
	t.Parallel()

	r := mustBuildSpec(t, &Config{
		PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
			injectAnnotations(t, g, "Pet.categories", WithEdgeUpdateBulk(true))
			return nil
		},
	})

	// Ensure create and update on categories have the bulk field in the right place.
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.categories`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.add_categories`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.remove_categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_categories`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_categories`))

	// And ensure it's not on others.
	assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.add_friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetCreate.properties.remove_friends`))
	assert.Nil(t, r.json(`$.components.schemas.PetUpdate.properties.friends`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.add_friends`))
	assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.remove_friends`))
}

func TestAnnotation_EdgesInUpsert(t *testing.T) {
	t.Parallel()

	t.Run("optional-edge-in-upsert", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Add upsert operation to Pet and include owner edge in upsert
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				injectAnnotations(t, g, "Pet.owner",
					WithIncludeOperations(
						OperationCreate,
						OperationUpdate,
						OperationUpsert,
					),
				)
				return nil
			},
		})

		// Verify Upsert schema includes the owner edge
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert`), "PetUpsert schema should exist")
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert.properties.owner`), "PetUpsert should include owner edge")
		assert.Equal(t, "string", r.json(`$.components.schemas.PetUpsert.properties.owner.type`))

		// Verify consistency with Create and Update
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.owner`))
		assert.NotNil(t, r.json(`$.components.schemas.PetUpdate.properties.owner`))
	})

	t.Run("edge-in-create-or-replace", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationCreateOrReplace,
						OperationDelete,
						OperationList,
					),
				)
				injectAnnotations(t, g, "Pet.owner",
					WithIncludeOperations(
						OperationCreate,
						OperationUpdate,
						OperationCreateOrReplace,
					),
				)
				return nil
			},
		})

		// Verify Replace schema includes the owner edge
		assert.NotNil(t, r.json(`$.components.schemas.PetReplace`))
		assert.NotNil(t, r.json(`$.components.schemas.PetReplace.properties.owner`))
		assert.Equal(t, "string", r.json(`$.components.schemas.PetReplace.properties.owner.type`))
	})

	t.Run("non-unique-edge-in-upsert", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Category has a non-unique edge to Pet (pets)
				injectAnnotations(t, g, "Category",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				injectAnnotations(t, g, "Category.pets",
					WithIncludeOperations(
						OperationCreate,
						OperationUpdate,
						OperationUpsert,
					),
				)
				return nil
			},
		})

		// Verify Upsert schema includes the pets edge as an array
		assert.NotNil(t, r.json(`$.components.schemas.CategoryUpsert`))
		assert.NotNil(t, r.json(`$.components.schemas.CategoryUpsert.properties.pets`))
		assert.Equal(t, "array", r.json(`$.components.schemas.CategoryUpsert.properties.pets.type`))
		assert.Equal(t, "integer", r.json(`$.components.schemas.CategoryUpsert.properties.pets.items.type`))
	})

	t.Run("edge-excluded-from-upsert", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				// Explicitly exclude owner from upsert
				injectAnnotations(t, g, "Pet.owner",
					WithExcludeOperations(OperationUpsert),
				)
				return nil
			},
		})

		// Verify Upsert schema exists but does NOT include the owner edge
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert`))
		assert.Nil(t, r.json(`$.components.schemas.PetUpsert.properties.owner`))

		// But it should still be in Create
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.owner`))
	})
}

func TestAnnotation_EdgeOperationInheritance(t *testing.T) {
	t.Parallel()

	t.Run("edge-inherits-upsert-from-entity", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Add upsert operation to Pet entity only, NOT to the owner edge
				// The edge should automatically inherit this from the entity
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				// Note: No annotation on Pet.owner - it should inherit from Pet
				return nil
			},
		})

		// Verify Upsert schema includes the owner edge even though edge wasn't annotated
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert`), "PetUpsert schema should exist")
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert.properties.owner`), "PetUpsert should include owner edge via inheritance")
		assert.Equal(t, "string", r.json(`$.components.schemas.PetUpsert.properties.owner.type`))
	})

	t.Run("edge-inherits-replace-from-entity", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Add replace operation to Pet entity only
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationCreateOrReplace,
						OperationDelete,
						OperationList,
					),
				)
				// Note: No annotation on Pet.owner - it should inherit from Pet
				return nil
			},
		})

		// Verify Replace schema includes the owner edge via inheritance
		assert.NotNil(t, r.json(`$.components.schemas.PetReplace`), "PetReplace schema should exist")
		assert.NotNil(t, r.json(`$.components.schemas.PetReplace.properties.owner`), "PetReplace should include owner edge via inheritance")
		assert.Equal(t, "string", r.json(`$.components.schemas.PetReplace.properties.owner.type`))
	})

	t.Run("edge-inherits-create-when-not-in-defaults", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			// Override default operations to NOT include Create
			DefaultOperations: []Operation{OperationRead, OperationList},
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Pet entity explicitly enables Create
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationList,
					),
				)
				// Note: No annotation on Pet.owner - it should inherit Create from Pet, not from defaults
				return nil
			},
		})

		// Verify Create schema includes the owner edge via entity inheritance (not defaults)
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate`), "PetCreate schema should exist")
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.owner`), "PetCreate should include owner edge via inheritance from entity")
	})

	t.Run("edge-exclusion-overrides-entity-inheritance", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Entity enables upsert
				injectAnnotations(t, g, "Pet",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				// Edge explicitly excludes upsert - this should override entity's operations
				injectAnnotations(t, g, "Pet.owner",
					WithExcludeOperations(OperationUpsert),
				)
				return nil
			},
		})

		// Verify Upsert schema exists but does NOT include the owner edge
		assert.NotNil(t, r.json(`$.components.schemas.PetUpsert`))
		assert.Nil(t, r.json(`$.components.schemas.PetUpsert.properties.owner`), "owner edge should be excluded due to WithExcludeOperations")

		// But Create should still have the edge (excluded only upsert)
		assert.NotNil(t, r.json(`$.components.schemas.PetCreate.properties.owner`))
	})

	t.Run("non-unique-edge-inherits-from-entity", func(t *testing.T) {
		t.Parallel()

		r := mustBuildSpec(t, &Config{
			PreGenerateHook: func(g *gen.Graph, _ *ogen.Spec) error {
				// Category has a non-unique edge to Pet (pets)
				injectAnnotations(t, g, "Category",
					WithIncludeOperations(
						OperationCreate,
						OperationRead,
						OperationUpdate,
						OperationUpsert,
						OperationDelete,
						OperationList,
					),
				)
				// Note: No annotation on Category.pets - it should inherit from Category
				return nil
			},
		})

		// Verify Upsert schema includes the pets edge as an array via inheritance
		assert.NotNil(t, r.json(`$.components.schemas.CategoryUpsert`))
		assert.NotNil(t, r.json(`$.components.schemas.CategoryUpsert.properties.pets`), "CategoryUpsert should include pets edge via inheritance")
		assert.Equal(t, "array", r.json(`$.components.schemas.CategoryUpsert.properties.pets.type`))
		assert.Equal(t, "integer", r.json(`$.components.schemas.CategoryUpsert.properties.pets.items.type`))
	})
}
