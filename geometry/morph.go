// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geometry

import (
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/math32"
	"strconv"
	"sort"
)

// MorphGeometry represents a base geometry and its morph targets.
type MorphGeometry struct {
	baseGeometry  *Geometry   // The base geometry
	targets       []*Geometry // The morph target geometries (containing deltas)
	weights       []float32   // The weights for each morph target
	uniWeights    gls.Uniform // Texture unit uniform location cache
	morphGeom     *Geometry   // Cache of the last CPU-morphed geometry
}

// NumMorphInfluencers is the maximum number of active morph targets.
const NumMorphTargets = 8

// NewMorphGeometry creates and returns a pointer to a new MorphGeometry.
func NewMorphGeometry(baseGeometry *Geometry) *MorphGeometry {

	mg := new(MorphGeometry)
	mg.baseGeometry = baseGeometry

	mg.targets = make([]*Geometry, 0)
	mg.weights = make([]float32, 0)

	mg.baseGeometry.ShaderDefines.Set("MORPHTARGETS", strconv.Itoa(NumMorphTargets))
	mg.uniWeights.Init("morphTargetInfluences")
	return mg
}

// GetGeometry satisfies the IGeometry interface.
func (mg *MorphGeometry) GetGeometry() *Geometry {

	return mg.baseGeometry
}

// SetWeights sets the morph target weights.
func (mg *MorphGeometry) SetWeights(weights []float32) {

	if len(weights) != len(mg.weights) {
		panic("weights have invalid length")
	}
	mg.weights = weights
}

// Weights returns the morph target weights.
func (mg *MorphGeometry) Weights() []float32 {

	return mg.weights
}

// AddMorphTargets add multiple morph targets to the morph geometry.
// Morph target deltas are calculated internally and the morph target geometries are altered to hold the deltas instead.
func (mg *MorphGeometry) AddMorphTargets(morphTargets ...*Geometry) {

	for i := range morphTargets {
		mg.weights = append(mg.weights, 0)
		// Calculate deltas for VertexPosition
		vertexIdx := 0
		baseVerticesVbo := mg.baseGeometry.VBO(gls.VertexPosition)
		baseVerticesStride := baseVerticesVbo.Stride()
		baseVerticesOffset := baseVerticesVbo.AttribOffset(gls.VertexPosition)
		baseVertices := baseVerticesVbo.Buffer()
		morphTargets[i].OperateOnVertices(func(vertex *math32.Vector3) bool {
			var baseVertex math32.Vector3
			//vertexIdx * baseVerticesStride + baseVerticesOffset
			baseVertices.GetVector3(vertexIdx * baseVerticesStride + baseVerticesOffset, &baseVertex)
			vertex.Sub(&baseVertex)
			vertexIdx++
			return false
		})
		// Calculate deltas for VertexNormal if attribute is present in target geometry
		// It is assumed that if VertexNormals are present in a target geometry then they are also present in the base geometry
		normalIdx := 0
		baseNormalsVBO := mg.baseGeometry.VBO(gls.VertexNormal)
		if baseNormalsVBO != nil {
			baseNormals := baseNormalsVBO.Buffer()
			baseNormalsStride := baseNormalsVBO.Stride()
			baseNormalsOffset := baseNormalsVBO.AttribOffset(gls.VertexNormal)
			morphTargets[i].OperateOnVertexNormals(func(normal *math32.Vector3) bool {
				var baseNormal math32.Vector3
				baseNormals.GetVector3(normalIdx * baseNormalsStride + baseNormalsOffset, &baseNormal)
				normal.Sub(&baseNormal)
				normalIdx++
				return false
			})
		}
		// TODO Calculate deltas for VertexTangents
	}
	mg.targets = append(mg.targets, morphTargets...)

	// Update all target attributes if we have few enough that we are able to send them
	// all to the shader without sorting and choosing the ones with highest current weight
	if len(mg.targets) <= NumMorphTargets {
		mg.UpdateTargetAttributes(mg.targets)
	}

}

// AddMorphTargetDeltas add multiple morph target deltas to the morph geometry.
func (mg *MorphGeometry) AddMorphTargetDeltas(morphTargetDeltas ...*Geometry) {

	for range morphTargetDeltas {
		mg.weights = append(mg.weights, 0)
	}
	mg.targets = append(mg.targets, morphTargetDeltas...)

	// Update all target attributes if we have few enough that we are able to send them
	// all to the shader without sorting and choosing the ones with highest current weight
	if len(mg.targets) <= NumMorphTargets {
		mg.UpdateTargetAttributes(mg.targets)
	}
}

// ActiveMorphTargets sorts the morph targets by weight and returns the top n morph targets with largest weight.
func (mg *MorphGeometry) ActiveMorphTargets() ([]*Geometry, []float32) {

	numTargets := len(mg.targets)
	if numTargets == 0 {
		return nil, nil
	}

	if numTargets <= NumMorphTargets {
		// No need to sort - just return the targets and weights directly
		return mg.targets, mg.weights
	} else {
		// Need to sort them by weight and only return the top N morph targets with largest weight (N = NumMorphTargets)
		// TODO test this (more than [NumMorphTargets] morph targets)
		sortedMorphTargets := make([]*Geometry, numTargets)
		copy(sortedMorphTargets, mg.targets)
		sort.Slice(sortedMorphTargets, func(i, j int) bool {
			return mg.weights[i] > mg.weights[j]
		})

		sortedWeights := make([]float32, numTargets)
		copy(sortedWeights, mg.weights)
		sort.Slice(sortedWeights, func(i, j int) bool {
			return mg.weights[i] > mg.weights[j]
		})
		return sortedMorphTargets, sortedWeights
	}
}

// SetIndices sets the indices array for this geometry.
func (mg *MorphGeometry) SetIndices(indices math32.ArrayU32) {

	mg.baseGeometry.SetIndices(indices)
	for i := range mg.targets {
		mg.targets[i].SetIndices(indices)
	}
}

// ComputeMorphed computes a morphed geometry from the provided morph target weights.
// Note that morphing is usually computed by the GPU in shaders.
// This CPU implementation allows users to obtain an instance of a morphed geometry
// if so desired (loosing morphing ability).
func (mg *MorphGeometry) ComputeMorphed(weights []float32) *Geometry {

	morphed := NewGeometry()
	// TODO
	return morphed
}

// Dispose releases, if possible, OpenGL resources, C memory
// and VBOs associated with the base geometry and morph targets.
func (mg *MorphGeometry) Dispose() {

	mg.baseGeometry.Dispose()
	for i := range mg.targets {
		mg.targets[i].Dispose()
	}
}

// UpdateTargetAttributes updates the attribute names of the specified morph targets in order.
func (mg *MorphGeometry) UpdateTargetAttributes(morphTargets []*Geometry) {

	for i, mt := range morphTargets {
		mt.SetAttributeName(gls.VertexPosition, "MorphPosition"+strconv.Itoa(i))
		mt.SetAttributeName(gls.VertexNormal, "MorphNormal"+strconv.Itoa(i))
		mt.SetAttributeName(gls.VertexTangent, "MorphTangent"+strconv.Itoa(i))
	}
}

// RenderSetup is called by the renderer before drawing the geometry.
func (mg *MorphGeometry) RenderSetup(gs *gls.GLS) {

	mg.baseGeometry.RenderSetup(gs)

	// Sort weights and find top 8 morph targets with largest current weight (8 is the max sent to shader)
	activeMorphTargets, activeWeights := mg.ActiveMorphTargets()

	// If the morph geometry has more targets than the shader supports we need to update attribute names
	// as weights change - we only send the top morph targets with highest weight
	if len(mg.targets) > NumMorphTargets {
		mg.UpdateTargetAttributes(activeMorphTargets)
	}

	// Transfer morphed geometry VBOs
	for _, mt := range activeMorphTargets {
		for _, vbo := range mt.VBOs() {
			vbo.Transfer(gs)
		}
	}

	mg.uniWeights.Uniform1fv(gs, int32(len(activeWeights)), &activeWeights[0])
}
