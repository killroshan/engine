
#if defined SKINNED

    vec4 transformed_temp;
    transformed_temp = (BindMatrix * vec4(transformed, 1));
    transformed_temp = Weights.x * BoneMatrices[int(Joints.x)] * transformed_temp +
                  Weights.y * BoneMatrices[int(Joints.y)] * transformed_temp +
                  Weights.z * BoneMatrices[int(Joints.z)] * transformed_temp +
                  Weights.w * BoneMatrices[int(Joints.w)] * transformed_temp;
    transformed_temp = BindMatrixInverse * transformed_temp;
    transformed = transformed_temp.xyz;

#endif
