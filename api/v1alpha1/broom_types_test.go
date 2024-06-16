package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestAdjustMemory(t *testing.T) {
	tests := map[string]struct {
		beforeMemoryLimit string
		adjustment        BroomAdjustment
		afterMemoryLimit  string
	}{
		"add adjustment": {
			beforeMemoryLimit: "100Mi",
			adjustment:        BroomAdjustment{Type: AddAdjustment, Value: "100Mi"},
			afterMemoryLimit:  "200Mi",
		},
		"mul adjustment": {
			beforeMemoryLimit: "100Mi",
			adjustment:        BroomAdjustment{Type: MulAdjustment, Value: "2"},
			afterMemoryLimit:  "200Mi",
		},
		"max limit reached with add adjustment": {
			beforeMemoryLimit: "100Mi",
			adjustment:        BroomAdjustment{Type: AddAdjustment, Value: "100Mi", MaxLimit: resource.MustParse("150Mi")},
			afterMemoryLimit:  "150Mi",
		},
		"max limit reached with mul adjustment": {
			beforeMemoryLimit: "100Mi",
			adjustment:        BroomAdjustment{Type: MulAdjustment, Value: "2", MaxLimit: resource.MustParse("150Mi")},
			afterMemoryLimit:  "150Mi",
		},
		"already at max limit": {
			beforeMemoryLimit: "100Mi",
			adjustment:        BroomAdjustment{Type: AddAdjustment, Value: "100Mi", MaxLimit: resource.MustParse("100Mi")},
			afterMemoryLimit:  "100Mi",
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := resource.MustParse(test.beforeMemoryLimit)
			adj := test.adjustment
			err := adj.AdjustMemory(&m)
			assert.Nil(t, err)
			assert.True(t, m.Equal(resource.MustParse(test.afterMemoryLimit)))
		})
	}
}
