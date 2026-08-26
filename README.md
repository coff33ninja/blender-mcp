# blender-mcp (Go)

A Go implementation of an MCP server for Blender.

**Source / credits:** the MCP protocol and tool info this is built on came from studying Blender Lab's [`MCP` add-on](https://www.blender.org/lab/mcp-server/#mcp-server) (bundled in this repo as `mcp-1.0.0/`) — a plugin that adds MCP socket support to Blender, not a full MCP server on its own. [djeada/blender-mcp-server](https://github.com/djeada/blender-mcp-server), an independent Python MCP server for Blender, was also referenced during development.

## Overview

This is a Go implementation of the Model Context Protocol (MCP) server that bridges MCP clients to Blender's TCP socket server. The `mcp` add-on (`mcp-1.0.0/`) runs inside Blender and exposes a TCP socket — this Go binary speaks MCP over stdio and forwards tool calls to Blender over that socket.

## Requirements

- **Blender 5.1+** with the `mcp` add-on installed and the socket server running
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

Add to your MCP client config (e.g. `opencode.json` or Claude Desktop config):

```json
{
  "mcpServers": {
    "blender": {
      "command": "/path/to/blender-mcp.exe",
      "args": ["--host", "localhost", "--port", "9876"]
    }
  }
}
```

## Exposed Tools

| Tool | Description |
|------|-------------|
| **Scene** | |
| `get_scene_info` | Scene metadata — name, camera, frame range, render engine, resolution |
| `list_objects` | List objects with optional type filter |
| `get_object_info` | Detailed properties of a specific object |
| `get_transform` | Get position, rotation, and scale of an object |
| `get_object_hierarchy` | Parent/child tree of an object |
| `list_collections` | List all collections and their object counts |
| `manage_collections` | Create collections and move objects between them |
| `set_frame_range` | Set frame range, FPS, and jump to frame |
| **Object Manipulation** | |
| `create_mesh_object` | Create cube, sphere, plane, cylinder, cone, torus, or monkey (data API) |
| `delete_object` | Remove an object by name |
| `duplicate_object` | Duplicate an object |
| `set_object_location` | Move an object to world coordinates |
| `set_object_rotation` | Set an object's Euler rotation |
| `set_object_scale` | Set an object's scale |
| `apply_transforms` | Bake location/rotation/scale into mesh data |
| **Materials** | |
| `material_create` | Create a material with optional base color |
| `material_assign` | Assign an existing material to an object |
| `set_object_material` | Create+assign material in one step |
| `set_material_color` | Set base color of an existing material |
| `set_material_texture` | Assign an image texture to a material |
| `list_materials` | List all materials in the blend file |
| **Node Trees & Animation** | |
| `get_node_tree` | Read shader/compositor node tree |
| `get_animation_data` | Get keyframes, NLA strips, and drivers |
| `setup_keyframes` | Insert transform keyframes on objects at specified frames |
| **Modifiers, Constraints & Physics** | |
| `manage_modifiers` | List, add, or remove modifiers |
| `manage_constraints` | List, add, or remove constraints |
| `manage_physics` | List, add, or remove physics simulations |
| `setup_rigid_body` | Add rigid body physics to one or more objects |
| **Fluid Simulation** | |
| `setup_fluid_domain` | Create a Mantaflow fluid domain (liquid or gas) |
| `setup_fluid_inflow` | Add a fluid inflow (flow) object |
| `setup_effector` | Set up collision/effector objects for fluid simulations |
| **Rendering & Export** | |
| `render_scene` | Render the current frame to a file |
| `render_animation` | Render a frame range to disk |
| `get_viewport_screenshot` | Capture the 3D viewport as a base64 PNG |
| `set_viewport_shading` | Set viewport mode (wireframe/solid/material/rendered) |
| `export_3d_file` | Export to any supported format by extension |
| `export_gltf` | Export as glTF/GLB |
| `export_obj` | Export as OBJ |
| `export_fbx` | Export as FBX |
| `import_3d_file` | Import .glb, .gltf, .fbx, .obj, .stl, .ply |
| **View & Lights** | |
| `set_viewport_camera` | Position the 3D viewport camera |
| `add_light` | Add POINT, SUN, SPOT, or AREA light |
| `add_camera` | Add a camera to the scene |
| `setup_camera` | Create/configure a camera with focal length and position |
| **File & History** | |
| `save_blend` | Save the current .blend file |
| `undo` | Undo the last action |
| `redo` | Redo the last undone action |
| **Code Execution** | |
| `execute_code` | Run arbitrary Python code in Blender's context |

**48 tools total.** `execute_code` is always available as the escape hatch.

## Architecture

```
MCP Client (stdio) ←→ blender-mcp.exe ←TCP→ Blender add-on (localhost:9876)
```

- **MCP transport:** stdio with JSON-RPC 2.0
- **Blender transport:** TCP with null-byte-delimited JSON
- **Code execution:** All tools generate Python code sent to Blender's embedded interpreter via `execute_code`

## License

MIT — see [LICENSE](LICENSE).
