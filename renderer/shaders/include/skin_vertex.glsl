
#if defined SKINNED

    transformed = (BindMatrix * vec4(VertexPosition, 1));
    vec4 skinPos = vec4(0.0);

    transformed = Weights.x * BoneMatrices[int(Joints.x)] * transformed +
                  Weights.y * BoneMatrices[int(Joints.y)] * transformed +
                  Weights.z * BoneMatrices[int(Joints.z)] * transformed +
                  Weights.w * BoneMatrices[int(Joints.w)] * transformed;
    transformed = BindMatrixInverse * transformed;
#endif
