# blender-mcp (Go)

A Go MCP server that bridges MCP clients (via stdio) to Blender's TCP socket server.

**Source / credits:** Built on Blender Lab's [`MCP` add-on](https://www.blender.org/lab/mcp-server/#mcp-server) protocol. [djeada/blender-mcp-server](https://github.com/djeada/blender-mcp-server) was also referenced during development.

## Overview

This Go binary speaks MCP over stdio (JSON-RPC 2.0) and forwards tool calls to Blender over TCP. The `mcp` add-on runs inside Blender and exposes a socket server — this tool bridges the two.

## Requirements

- **Blender 5.1+** with the `mcp` add-on installed and socket server running on `localhost:9876`
- Go 1.26+ (for building from source)

## Building

```bash
cd go-blender-mcp
go build -o blender-mcp.exe .
```

## Usage

```bash
blender-mcp.exe --host localhost --port 9876 --verbose
```

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `localhost` | Blender TCP server host |
| `--port` | `9876` | Blender TCP server port |
| `--verbose` | `false` | Enable debug logging to stderr |

## MCP Client Configuration

```json
{
  "mcpServers": {
    "blender": {
      "command": "blender-mcp.exe",
      "args": ["--host", "localhost", "--port", "9876"]
    }
  }
}
```

## Exposed Tools (63)

### Scene & Inspection
| Tool | Description |
|------|-------------|
| `get_scene_info` | Scene metadata — name, camera, frame range, render engine, resolution, object count |
| `list_objects` | List objects with optional type filter (MESH, LIGHT, CAMERA, etc.) |
| `get_object_info` | Detailed properties of a specific object |
| `get_transform` | Get position, rotation, and scale of an object |
| `get_object_hierarchy` | Parent/child tree of an object with configurable depth |
| `list_collections` | List all collections and their object counts |
| `list_materials` | List all materials in the blend file |

### Object Creation & Manipulation
| Tool | Description |
|------|-------------|
| `create_mesh_object` | Create cube, sphere, plane, cylinder, cone, torus, or monkey (data API) |
| `create_object` | Create any Blender object type: EMPTY, ARMATURE, CURVE, CURVES, FONT, GREASEPENCIL, LATTICE, LIGHT_PROBE, META, POINTCLOUD, SPEAKER, SURFACE, VOLUME |
| `delete_object` | Delete an object by name |
| `duplicate_object` | Duplicate an existing object |
| `set_object_location` | Set the world-space location of an object |
| `set_object_rotation` | Set an object's Euler rotation (radians) |
| `set_object_scale` | Set an object's scale |
| `apply_transforms` | Bake location/rotation/scale into mesh data |
| `edit_mesh_data` | Direct mesh editing: add/remove vertices, edges, faces |

### Materials & Nodes
| Tool | Description |
|------|-------------|
| `material_create` | Create a material with optional base color |
| `material_assign` | Assign an existing material to an object |
| `set_object_material` | Create+assign material in one step with configurable color |
| `set_material_color` | Set base color of an existing material's Principled BSDF |
| `set_material_texture` | Assign an image texture to a material |
| `get_node_tree` | Read shader/compositor node tree of an object's material |

### Modifiers, Constraints & Physics
| Tool | Description |
|------|-------------|
| `manage_modifiers` | List, add, or remove any modifier (SUBSURF, MIRROR, ARRAY, BOOLEAN, ARMATURE, CLOTH, FLUID, etc.) |
| `manage_constraints` | List, add, or remove any constraint (TRACK_TO, COPY_LOCATION, IK, CHILD_OF, FOLLOW_PATH, etc.) |
| `manage_physics` | List, add, or remove physics (RIGID_BODY, CLOTH, FLUID, FORCE_FIELD, SOFT_BODY, PARTICLE_SYSTEM, DYNAMIC_PAINT) |
| `manage_shader_nodes` | List, add, remove, or connect nodes in a material's shader node tree |

### Fluid Simulation
| Tool | Description |
|------|-------------|
| `setup_fluid_domain` | Create a Mantaflow fluid domain (liquid or gas) |
| `setup_fluid_inflow` | Add a fluid inflow (flow) object |
| `setup_effector` | Set up collision/effector objects for fluid simulations |

### Animation & Rigging
| Tool | Description |
|------|-------------|
| `setup_keyframes` | Insert transform keyframes on objects at specified frames |
| `setup_rigid_body` | Add rigid body physics to one or more objects |
| `set_particle_system` | Add or configure a particle system (emitter or hair) |
| `get_animation_data` | Get keyframes, NLA strips, and drivers |
| `rigify_ops` | Rigify operations: generate rig, add metarig (14 types), list rig types |
| `pose_library_ops` | Pose library: save/apply poses as actions, list pose actions, create library |

### Rendering
| Tool | Description |
|------|-------------|
| `render_scene` | Render the current scene to a file |
| `render_animation` | Render a frame range to disk |
| `get_viewport_screenshot` | Capture the 3D viewport as a base64 PNG |
| `set_viewport_shading` | Set viewport mode (wireframe/solid/material/rendered) |
| `set_viewport_camera` | Position the 3D viewport camera |
| `set_render_engine` | Switch render engine (BLENDER_EEVEE, BLENDER_CYCLES, HYDRA_STORM) |
| `set_render_format` | Set output format, resolution, color mode, and color depth |
| `set_render_passes` | Enable/disable render passes (30 types: diffuse, glossy, emission, AO, normal, Z, cryptomatte, etc.) |

### Export & Import
| Tool | Description |
|------|-------------|
| `export_3d_file` | Export to any supported format by extension (.glb, .gltf, .fbx, .obj, .stl, .ply, .bvh, .svg) |
| `export_gltf` | Export as glTF/GLB |
| `export_obj` | Export as OBJ |
| `export_fbx` | Export as FBX |
| `import_3d_file` | Import .glb, .gltf, .fbx, .obj, .stl, .ply, .bvh, .svg |

### Lights, Cameras & Environment
| Tool | Description |
|------|-------------|
| `add_light` | Add POINT, SUN, SPOT, or AREA light with color and energy |
| `add_camera` | Add a camera to the scene |
| `setup_camera` | Create/configure a camera with focal length and position |
| `set_world_environment` | Configure world background color, strength, or environment texture |
| `set_cursor` | Set the 3D cursor location and rotation |
| `set_snap` | Configure snap settings: type, target, toggle |

### Collections & File Management
| Tool | Description |
|------|-------------|
| `manage_collections` | Create collections, move objects between them, set parent/color tags |
| `save_blend` | Save the current .blend file |
| `undo` | Undo the last action |
| `redo` | Redo the last undone action |

### Grease Pencil (28 actions)
| Tool | Description |
|------|-------------|
| `grease_pencil_manage` | Full GP management: layers, modifiers (25 types), strokes, frames, fill, extrude, dissolve, duplicate, clean, convert, and more |

### Compositor & Node Wrangling
| Tool | Description |
|------|-------------|
| `compositor_nodes` | Manage compositor nodes: list types (90+), list nodes, add, remove, connect |
| `node_wrangler_ops` | Node Wrangler shortcuts: preview, swap, mix, collapse/expand, frame, texture setup, connect viewer |

> **Note:** `node_wrangler_ops` requires the Shader Editor to be the active context in Blender. This means it won't work in headless/background mode — only when Blender's UI is open with the Shader Editor visible.

### Code Execution
| Tool | Description |
|------|-------------|
| `execute_code` | Run arbitrary Python code in Blender's context (escape hatch for anything not covered) |

## Architecture

```
MCP Client (stdio JSON-RPC 2.0) ←→ blender-mcp.exe ←TCP (null-byte JSON)→ Blender add-on (localhost:9876)
```

- **MCP transport:** stdio with JSON-RPC 2.0
- **Blender transport:** TCP with null-byte-delimited JSON, per-request connection
- **All tools** generate Python code sent to Blender's embedded interpreter via `execute_code`

## License

MIT — see [LICENSE](LICENSE).
