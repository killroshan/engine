//
// Physically Based Shading of a microfacet surface material - Vertex Shader
// Modified from reference implementation at https://github.com/KhronosGroup/glTF-WebGL-PBR
//
#include <attributes>

// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

// skinned mesh
uniform mat4 BindMatrix;
uniform mat4 BindMatrixInverse;

#if defined SKINNED
    uniform mat4 BoneMatrices[MAX_BONES];
#endif

#include <morphtarget_vertex_declaration>

// Output variables for Fragment shader
out vec3 Position;
out vec3 Normal;
out vec3 CamDir;
out vec2 FragTexcoord;

void main() {

    // Transform this vertex position to camera coordinates.
    Position = vec3(ModelViewMatrix * vec4(VertexPosition, 1.0));

    // Transform this vertex normal to camera coordinates.
    Normal = normalize(NormalMatrix * VertexNormal);

    // Calculate the direction vector from the vertex to the camera
    // The camera is at 0,0,0
    CamDir = normalize(-Position.xyz);

    // Flips texture coordinate Y if requested.
    vec2 texcoord = VertexTexcoord;
    // #if MAT_TEXTURES>0
    //     if (MatTexFlipY(0)) {
    //         texcoord.y = 1 - texcoord.y;
    //     }
    // #endif
    FragTexcoord = texcoord;

    vec4 transformed = (BindMatrix * vec4(VertexPosition, 1));
    vec4 skinPos = vec4(0.0);

    transformed = Weights.x * BoneMatrices[int(Joints.x)] * transformed +
                  Weights.y * BoneMatrices[int(Joints.y)] * transformed +
                  Weights.z * BoneMatrices[int(Joints.z)] * transformed +
                  Weights.w * BoneMatrices[int(Joints.w)] * transformed;


    transformed = BindMatrixInverse * transformed;

    #include <morphtarget_vertex> [MORPHTARGETS]

    gl_Position = MVP * transformed;
}


