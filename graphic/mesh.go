// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graphic

import (
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

type Skeleton struct {
	Name                string
	Root                core.INode
	Bones               []core.INode
	BoneMatricesInverse []math32.Matrix4
	BoneMatrices        []math32.Matrix4
	uniBoneMatrices     gls.Uniform
}

func NewSkeleton(name string, root core.INode, bones []core.INode, matricesInverse []math32.Matrix4) *Skeleton {
	if len(bones) != len(matricesInverse) {
		panic("bones and matricesinverse mismatch")
	}

	bmi := make([]math32.Matrix4, 0, len(bones))
	bmi = append(bmi, matricesInverse...)

	return &Skeleton{
		Name:                name,
		Bones:               bones,
		Root:                root,
		BoneMatricesInverse: bmi,
		BoneMatrices:        make([]math32.Matrix4, len(bones), len(bones)),
	}
}

func (s *Skeleton) Clone() *Skeleton {
	bmi := make([]math32.Matrix4, 0, len(s.Bones))
	bmi = append(bmi, s.BoneMatricesInverse...)
	return &Skeleton{
		Name:                s.Name,
		Bones:               s.Bones,
		BoneMatricesInverse: bmi,
		BoneMatrices:        make([]math32.Matrix4, len(s.Bones), len(s.Bones)),
	}
}

func (s *Skeleton) update() {
	for i, bone := range s.Bones {
		var mat math32.Matrix4
		mat = bone.GetNode().MatrixWorld()
		s.BoneMatrices[i].MultiplyMatrices(&mat, &s.BoneMatricesInverse[i])
	}
}

func (s *Skeleton) pose() {
	for idx, bone := range s.Bones {
		var mat math32.Matrix4
		mat.GetInverse(&s.BoneMatricesInverse[idx])
		bone.GetNode().SetMatrixWorld(&mat)
	}

	for _, bone := range s.Bones {
		parent := bone.GetNode().Parent()
		if parent != nil && parent.GetNode().IsBone == true {
			var mat math32.Matrix4
			parent_mat_world := parent.GetNode().MatrixWorld()
			mat_world := bone.GetNode().MatrixWorld()
			mat.GetInverse(&parent_mat_world)
			mat.Multiply(&mat_world)
			bone.GetNode().SetMatrix(&mat)
			bone.GetNode().SetChanged(true)
		} else {
			mat := bone.GetNode().MatrixWorld()
			bone.GetNode().SetMatrix(&mat)
			bone.GetNode().SetChanged(true)
		}
	}

}

// Mesh is a Graphic with uniforms for the model, view, projection, and normal matrices.
type Mesh struct {
	Graphic             // Embedded graphic
	uniMVm  gls.Uniform // Model view matrix uniform location cache
	uniMVPm gls.Uniform // Model view projection matrix uniform cache
	uniNm   gls.Uniform // Normal matrix uniform cache
}

// mesh which support skinned animation
type SkinnedMesh struct {
	Mesh
	Skeleton          *Skeleton
	BindMatrix        math32.Matrix4
	BindMatrixInverse math32.Matrix4

	uniBind        gls.Uniform // Bind Matrix
	uniBindInverse gls.Uniform // Bind Matrix Inverse
}

func NewSkinnedMesh(igeom geometry.IGeometry, imat material.IMaterial) *SkinnedMesh {
	sm := new(SkinnedMesh)
	sm.Init(igeom, imat)
	sm.uniBind.Init("BindMatrix")
	sm.uniBindInverse.Init("BindMatrixInverse")
	return sm
}

func (sm *SkinnedMesh) Pose() {
	sm.Skeleton.pose()
}

func (sm *SkinnedMesh) Bind(skeleton *Skeleton) {
	sm.BindMatrix = sm.MatrixWorld()
	sm.BindMatrixInverse.GetInverse(&sm.BindMatrix)
	sm.Skeleton = skeleton
	skeleton.uniBoneMatrices.Init("BoneMatrices")
}

func (sm *SkinnedMesh) AddGroupMaterial(imat material.IMaterial, gindex int) {
	sm.Graphic.AddGroupMaterial(sm, imat, gindex)
}

func (sm *SkinnedMesh) AddMaterial(imat material.IMaterial, start, count int) {
	sm.Graphic.AddMaterial(sm, imat, start, count)
}

func (sm *SkinnedMesh) RenderSetup(gs *gls.GLS, rinfo *core.RenderInfo) {
	sm.Mesh.RenderSetup(gs, rinfo)

	sm.Skeleton.update()
	sm.uniBind.UniformMatrix4fv(gs, 1, false, &sm.BindMatrix[0])
	sm.uniBindInverse.UniformMatrix4fv(gs, 1, false, &sm.BindMatrixInverse[0])
	sm.Skeleton.uniBoneMatrices.UniformMatrix4fv(gs, int32(sm.MaxBones()), false, &sm.Skeleton.BoneMatrices[0][0])

}

func (sm *SkinnedMesh) UpdateMatrixWorld() {
	sm.GetNode().UpdateMatrixWorld()
	temp := sm.MatrixWorld()
	sm.BindMatrixInverse.GetInverse(&temp)
}

func (sm *SkinnedMesh) MaxBones() int {
	return len(sm.Skeleton.Bones)
}

func (sm *SkinnedMesh) NormalizeSkinWeights() {
	sm.GetGeometry().OperateOnSkinWeights(func(weights *math32.Vector4) bool {
		length := weights.ManhattanLength()
		if length == 0 {
			weights.Set(1, 0, 0, 0)
		} else {
			weights.MultiplyScalar(1 / length)
		}
		return false
	})
}

// NewMesh creates and returns a pointer to a mesh with the specified geometry and material.
// If the mesh has multi materials, the material specified here must be nil and the
// individual materials must be add using "AddMaterial" or AddGroupMaterial".
func NewMesh(igeom geometry.IGeometry, imat material.IMaterial) *Mesh {

	m := new(Mesh)
	m.Init(igeom, imat)
	return m
}

// Init initializes the Mesh and its uniforms.
func (m *Mesh) Init(igeom geometry.IGeometry, imat material.IMaterial) {

	m.Graphic.Init(igeom, gls.TRIANGLES)

	// Initialize uniforms
	m.uniMVm.Init("ModelViewMatrix")
	m.uniMVPm.Init("MVP")
	m.uniNm.Init("NormalMatrix")

	// Adds single material if not nil
	if imat != nil {
		m.AddMaterial(imat, 0, 0)
	}
}

// AddMaterial adds a material for the specified subset of vertices.
func (m *Mesh) AddMaterial(imat material.IMaterial, start, count int) {

	m.Graphic.AddMaterial(m, imat, start, count)
}

// AddGroupMaterial adds a material for the specified geometry group.
func (m *Mesh) AddGroupMaterial(imat material.IMaterial, gindex int) {

	m.Graphic.AddGroupMaterial(m, imat, gindex)
}

// RenderSetup is called by the engine before drawing the mesh geometry
// It is responsible to updating the current shader uniforms with
// the model matrices.
func (m *Mesh) RenderSetup(gs *gls.GLS, rinfo *core.RenderInfo) {

	// Transfer uniform for model view matrix
	mvm := m.ModelViewMatrix()
	m.uniMVm.UniformMatrix4fv(gs, 1, false, &mvm[0])

	// Transfer uniform for model view projection matrix
	mvpm := m.ModelViewProjectionMatrix()
	m.uniMVPm.UniformMatrix4fv(gs, 1, false, &mvpm[0])

	// Calculates normal matrix and transfer uniform
	var nm math32.Matrix3
	nm.GetNormalMatrix(mvm)
	m.uniNm.UniformMatrix3fv(gs, 1, false, &nm[0])
}

// Raycast checks intersections between this geometry and the specified raycaster
// and if any found appends it to the specified intersects array.
func (m *Mesh) Raycast(rc *core.Raycaster, intersects *[]core.Intersect) {

	// Transform this mesh geometry bounding sphere from model
	// to world coordinates and checks intersection with raycaster
	geom := m.GetGeometry()
	sphere := geom.BoundingSphere()
	matrixWorld := m.MatrixWorld()
	sphere.ApplyMatrix4(&matrixWorld)
	if !rc.IsIntersectionSphere(&sphere) {
		return
	}

	// Copy ray and transform to model coordinates
	// This ray will will also be used to check intersects with
	// the geometry, as is much less expensive to transform the
	// ray to model coordinates than the geometry to world coordinates.
	var inverseMatrix math32.Matrix4
	inverseMatrix.GetInverse(&matrixWorld)
	var ray math32.Ray
	ray.Copy(&rc.Ray).ApplyMatrix4(&inverseMatrix)
	bbox := geom.BoundingBox()
	if !ray.IsIntersectionBox(&bbox) {
		return
	}

	// Local function to check the intersection of the ray from the raycaster with
	// the specified face defined by three poins.
	checkIntersection := func(mat *material.Material, pA, pB, pC, point *math32.Vector3) *core.Intersect {

		var intersect bool
		switch mat.Side() {
		case material.SideBack:
			intersect = ray.IntersectTriangle(pC, pB, pA, true, point)
		case material.SideFront:
			intersect = ray.IntersectTriangle(pA, pB, pC, true, point)
		case material.SideDouble:
			intersect = ray.IntersectTriangle(pA, pB, pC, false, point)
		}
		if !intersect {
			return nil
		}

		// Transform intersection point from model to world coordinates
		var intersectionPointWorld = *point
		intersectionPointWorld.ApplyMatrix4(&matrixWorld)

		// Calculates the distance from the ray origin to intersection point
		origin := rc.Ray.Origin()
		distance := origin.DistanceTo(&intersectionPointWorld)

		// Checks if distance is between the bounds of the raycaster
		if distance < rc.Near || distance > rc.Far {
			return nil
		}

		return &core.Intersect{
			Distance: distance,
			Point:    intersectionPointWorld,
			Object:   m,
		}
	}

	i := 0
	geom.ReadFaces(func(vA, vB, vC math32.Vector3) bool {
		// Checks intersection of the ray with this face
		mat := m.GetMaterial(i).GetMaterial()
		var point math32.Vector3
		intersect := checkIntersection(mat, &vA, &vB, &vC, &point)
		if intersect != nil {
			intersect.Index = uint32(i)
			*intersects = append(*intersects, *intersect)
		}
		i++
		return false
	})
}
