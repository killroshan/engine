//
// Physically Based Shading of a microfacet surface material - Vertex Shader
// Modified from reference implementation at https://github.com/KhronosGroup/glTF-WebGL-PBR
//
#include <attributes>

// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

#include <skin>

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

    FragTexcoord = texcoord;
    vec4 transformed = vec4(VertexPosition, 1);

    #include <skin_vertex>

    gl_Position = MVP * transformed;
}


