package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coff33ninja/blender-mcp/blender"
)

// ToolHandler is a function that processes a tool call's arguments.
type ToolHandler func(args map[string]any) (any, *Error)

// ToolRegistry holds registered tool handlers.
type ToolRegistry struct {
	tools    []Tool
	handlers map[string]ToolHandler
}

// NewToolRegistry creates an empty registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		handlers: make(map[string]ToolHandler),
	}
}

// Register adds a tool definition and its handler.
func (tr *ToolRegistry) Register(tool Tool, handler ToolHandler) {
	tr.tools = append(tr.tools, tool)
	tr.handlers[tool.Name] = handler
}

// List returns all registered tools.
func (tr *ToolRegistry) List() []Tool {
	return tr.tools
}

// Call invokes a tool by name. Returns the ToolCallResult.
func (tr *ToolRegistry) Call(name string, args map[string]any) (any, *Error) {
	h, ok := tr.handlers[name]
	if !ok {
		return nil, &Error{Code: CodeMethodNotFound, Message: "tool not found: " + name}
	}
	return h(args)
}

// ToolDefs returns the list of tools this server exposes.
func ToolDefs() []Tool {
	return []Tool{
		{
			Name:        "execute_code",
			Description: "Execute arbitrary Python code in Blender's context. The code must set a `result` dict variable. Use `bpy` for Blender Python API access.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{
						"type":        "string",
						"description": "Python code to execute inside Blender. Must assign `result = {...}` for return data.",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "get_scene_info",
			Description: "Get current scene information including name, camera, render engine, frame range, and object count.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_objects",
			Description: "List all objects in the current scene with their type, location, and visibility.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filter_type": map[string]any{
						"type":        "string",
						"description": "Optional filter by object type (MESH, CURVE, LIGHT, CAMERA, etc.)",
					},
				},
			},
		},
		{
			Name:        "get_object_info",
			Description: "Get detailed information about a specific object by name.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "create_mesh_object",
			Description: "Create a new mesh object (cube, sphere, plane, cylinder, cone, torus, monkey).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"primitive": map[string]any{
						"type":        "string",
						"description": "Primitive type",
						"enum":        []string{"cube", "sphere", "plane", "cylinder", "cone", "torus", "monkey"},
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Object name (optional)",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "Location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"primitive"},
			},
		},
		{
			Name:        "delete_object",
			Description: "Delete an object by name from the scene.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name to delete",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "set_object_location",
			Description: "Set the world-space location of an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "Location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"name", "location"},
			},
		},
		{
			Name:        "set_object_material",
			Description: "Assign or create a material on an object with configurable color.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name",
					},
					"material_name": map[string]any{
						"type":        "string",
						"description": "Material name (created if it doesn't exist)",
					},
					"color": map[string]any{
						"type":        "array",
						"description": "RGBA color [r, g, b, a], values 0-1",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"name", "material_name"},
			},
		},
		{
			Name:        "render_scene",
			Description: "Render the current scene to a file. Returns the file path.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"output_path": map[string]any{
						"type":        "string",
						"description": "Output file path (e.g. /tmp/render.png)",
					},
					"resolution_x": map[string]any{
						"type":        "integer",
						"description": "Render width in pixels",
					},
					"resolution_y": map[string]any{
						"type":        "integer",
						"description": "Render height in pixels",
					},
				},
			},
		},
		{
			Name:        "get_viewport_screenshot",
			Description: "Capture the current 3D viewport as a base64-encoded PNG image.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "import_3d_file",
			Description: "Import a 3D model file (.glb, .gltf, .fbx, .obj, .stl, .ply) into the scene.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Absolute path to the file to import",
					},
				},
				"required": []string{"filepath"},
			},
		},
		{
			Name:        "export_3d_file",
			Description: "Export the scene or selected objects to a file (.glb, .gltf, .fbx, .obj, .stl, .ply).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Absolute path for the exported file",
					},
					"use_selection": map[string]any{
						"type":        "boolean",
						"description": "Export only selected objects",
					},
				},
				"required": []string{"filepath"},
			},
		},
		{
			Name:        "set_viewport_camera",
			Description: "Set the 3D viewport camera position and rotation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "array",
						"description": "Camera location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
					"rotation": map[string]any{
						"type":        "array",
						"description": "Camera rotation [x, y, z] in Euler angles (radians)",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"location", "rotation"},
			},
		},
		{
			Name:        "add_light",
			Description: "Add a light source to the scene.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"light_type": map[string]any{
						"type":        "string",
						"description": "Light type",
						"enum":        []string{"POINT", "SUN", "SPOT", "AREA"},
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Light name (optional)",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "Location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
					"energy": map[string]any{
						"type":        "number",
						"description": "Light energy/intensity",
					},
					"color": map[string]any{
						"type":        "array",
						"description": "RGB color [r, g, b], values 0-1",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"light_type"},
			},
		},
		{
			Name:        "add_camera",
			Description: "Add a camera to the scene.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Camera name (optional)",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "Location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
					"rotation": map[string]any{
						"type":        "array",
						"description": "Rotation [x, y, z] in Euler angles (radians)",
						"items":       map[string]any{"type": "number"},
					},
				},
			},
		},
		{
			Name:        "set_object_rotation",
			Description: "Set the Euler rotation of an object in radians.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name",
					},
					"rotation": map[string]any{
						"type":        "array",
						"description": "Rotation [x, y, z] in radians",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"name", "rotation"},
			},
		},
		{
			Name:        "set_object_scale",
			Description: "Set the scale of an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name",
					},
					"scale": map[string]any{
						"type":        "array",
						"description": "Scale [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"name", "scale"},
			},
		},
		{
			Name:        "duplicate_object",
			Description: "Duplicate an existing object by name.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "The object name to duplicate",
					},
					"new_name": map[string]any{
						"type":        "string",
						"description": "Name for the duplicate (optional)",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "undo",
			Description: "Undo the last action in Blender.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "redo",
			Description: "Redo the last undone action in Blender.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_materials",
			Description: "List all materials in the blend file.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_collections",
			Description: "List all collections and their object counts.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "render_animation",
			Description: "Render an animation frame range to disk.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"output_dir": map[string]any{
						"type":        "string",
						"description": "Output directory (files will be named frame_NNNN.ext)",
					},
					"frame_start": map[string]any{
						"type":        "integer",
						"description": "Start frame (defaults to scene start)",
					},
					"frame_end": map[string]any{
						"type":        "integer",
						"description": "End frame (defaults to scene end)",
					},
				},
			},
		},
		{
			Name:        "get_node_tree",
			Description: "Get the shader/compositor/node tree of an object's material as a list of nodes and links.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object whose material nodes to read",
					},
					"material_slot": map[string]any{
						"type":        "integer",
						"description": "Material slot index (default 0)",
					},
				},
				"required": []string{"object_name"},
			},
		},
		{
			Name:        "get_animation_data",
			Description: "Get animation data for an object: keyframes, NLA strips, and drivers.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name (defaults to active object if omitted)",
					},
				},
			},
		},
		{
			Name:        "get_object_hierarchy",
			Description: "Get the parent/child hierarchy of an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Root object name (defaults to active object if omitted)",
					},
					"max_depth": map[string]any{
						"type":        "integer",
						"description": "Max traversal depth (default 3)",
					},
				},
			},
		},
		{
			Name:        "manage_modifiers",
			Description: "List, add, or remove modifiers on an object. action: 'list' (default), 'add', or 'remove'.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "list, add, or remove",
						"enum":        []string{"list", "add", "remove"},
					},
					"modifier_type": map[string]any{
						"type":        "string",
						"description": "Modifier type to add (e.g. SUBSURF, MIRROR, ARRAY, SOLIDIFY, BEVEL)",
					},
					"modifier_name": map[string]any{
						"type":        "string",
						"description": "Modifier name to remove",
					},
				},
				"required": []string{"object_name"},
			},
		},
		{
			Name:        "manage_constraints",
			Description: "List, add, or remove constraints on an object. action: 'list' (default), 'add', or 'remove'.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "list, add, or remove",
						"enum":        []string{"list", "add", "remove"},
					},
					"constraint_type": map[string]any{
						"type":        "string",
						"description": "Constraint type to add (e.g. TRACK_TO, COPY_LOCATION, COPY_ROTATION, LIMIT_DISTANCE, CHILD_OF)",
					},
					"constraint_name": map[string]any{
						"type":        "string",
						"description": "Constraint name to remove",
					},
					"target_name": map[string]any{
						"type":        "string",
						"description": "Target object for the constraint",
					},
				},
				"required": []string{"object_name"},
			},
		},
		{
			Name:        "manage_physics",
			Description: "List, add, or remove physics simulations on an object. action: 'list' (default), 'add', or 'remove'.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "list, add, or remove",
						"enum":        []string{"list", "add", "remove"},
					},
					"physics_type": map[string]any{
						"type":        "string",
						"description": "Physics type to add (e.g. RIGID_BODY, CLOTH, FLUID, FORCE_FIELD, SOFT_BODY, PARTICLE_SYSTEM)",
					},
				},
				"required": []string{"object_name"},
			},
		},
		{
			Name:        "set_material_color",
			Description: "Set the base color of an existing material's Principled BSDF node.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"material_name": map[string]any{
						"type":        "string",
						"description": "Material name",
					},
					"color": map[string]any{
						"type":        "array",
						"description": "RGBA color [r, g, b, a], values 0-1",
						"items":       map[string]any{"type": "number"},
					},
				},
				"required": []string{"material_name", "color"},
			},
		},
		{
			Name:        "set_material_texture",
			Description: "Assign an image texture to a material's Principled BSDF base color input.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"material_name": map[string]any{
						"type":        "string",
						"description": "Material name",
					},
					"image_path": map[string]any{
						"type":        "string",
						"description": "Absolute path to the image file",
					},
				},
				"required": []string{"material_name", "image_path"},
			},
		},
		{
			Name:        "get_transform",
			Description: "Get position, rotation, and scale of an object by name.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "material_create",
			Description: "Create a new material with optional base color.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Material name (default: Material)",
					},
					"color": map[string]any{
						"type":        "array",
						"description": "RGB or RGBA base color [r, g, b] or [r, g, b, a], values 0-1",
						"items":       map[string]any{"type": "number"},
					},
				},
			},
		},
		{
			Name:        "material_assign",
			Description: "Assign an existing material to an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
					"material_name": map[string]any{
						"type":        "string",
						"description": "Material name",
					},
				},
				"required": []string{"object_name", "material_name"},
			},
		},
		{
			Name:        "save_blend",
			Description: "Save the current .blend file.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Path to save to. Omit to save in place.",
					},
				},
			},
		},
		{
			Name:        "apply_transforms",
			Description: "Apply location, rotation, and/or scale transforms to objects.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_names": map[string]any{
						"type":        "array",
						"description": "Object names to apply transforms to",
						"items":       map[string]any{"type": "string"},
					},
					"location": map[string]any{
						"type":        "boolean",
						"description": "Apply location (default true)",
					},
					"rotation": map[string]any{
						"type":        "boolean",
						"description": "Apply rotation (default true)",
					},
					"scale": map[string]any{
						"type":        "boolean",
						"description": "Apply scale (default true)",
					},
				},
				"required": []string{"object_names"},
			},
		},
		{
			Name:        "set_frame_range",
			Description: "Set the scene frame range, FPS, and optionally jump to a frame.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"frame_start": map[string]any{
						"type":        "integer",
						"description": "Start frame",
					},
					"frame_end": map[string]any{
						"type":        "integer",
						"description": "End frame",
					},
					"frame_current": map[string]any{
						"type":        "integer",
						"description": "Jump to this frame",
					},
					"fps": map[string]any{
						"type":        "integer",
						"description": "Frames per second",
					},
				},
			},
		},
		{
			Name:        "set_viewport_shading",
			Description: "Set the 3D viewport shading mode.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mode": map[string]any{
						"type":        "string",
						"description": "Shading mode",
						"enum":        []string{"WIREFRAME", "SOLID", "MATERIAL", "RENDERED"},
					},
				},
				"required": []string{"mode"},
			},
		},
		{
			Name:        "export_gltf",
			Description: "Export the scene as glTF/GLB.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Output file path (.glb or .gltf)",
					},
				},
				"required": []string{"filepath"},
			},
		},
		{
			Name:        "export_obj",
			Description: "Export the scene as OBJ.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Output file path (.obj)",
					},
				},
				"required": []string{"filepath"},
			},
		},
		{
			Name:        "export_fbx",
			Description: "Export the scene as FBX.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filepath": map[string]any{
						"type":        "string",
						"description": "Output file path (.fbx)",
					},
				},
				"required": []string{"filepath"},
			},
		},
		{
			Name:        "setup_keyframes",
			Description: "Insert transform keyframes on objects at specified frames.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"keyframes": map[string]any{
						"type":        "array",
						"description": "List of keyframe specs: {object, frame, location, rotation, scale}",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"object":   map[string]any{"type": "string", "description": "Object name"},
								"frame":    map[string]any{"type": "integer", "description": "Frame number"},
								"location": map[string]any{"type": "array", "items": map[string]any{"type": "number"}, "description": "[x,y,z] location"},
								"rotation": map[string]any{"type": "array", "items": map[string]any{"type": "number"}, "description": "[rx,ry,rz] Euler rotation in radians"},
								"scale":    map[string]any{"type": "array", "items": map[string]any{"type": "number"}, "description": "[sx,sy,sz] scale"},
							},
							"required": []string{"object", "frame"},
						},
					},
				},
				"required": []string{"keyframes"},
			},
		},
		{
			Name:        "setup_rigid_body",
			Description: "Add rigid body physics to one or more objects.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_names": map[string]any{
						"type":        "array",
						"description": "Object names to configure",
						"items":       map[string]any{"type": "string"},
					},
					"rb_type": map[string]any{
						"type":        "string",
						"description": "Rigid body type",
						"enum":        []string{"ACTIVE", "PASSIVE"},
					},
					"mass": map[string]any{
						"type":        "number",
						"description": "Mass in kg (default 1.0)",
					},
					"friction": map[string]any{
						"type":        "number",
						"description": "Friction 0-1 (default 0.5)",
					},
					"restitution": map[string]any{
						"type":        "number",
						"description": "Bounciness 0-1 (default 0.3)",
					},
					"collision_shape": map[string]any{
						"type":        "string",
						"description": "Collision shape",
						"enum":        []string{"BOX", "SPHERE", "CAPSULE", "CYLINDER", "CONE", "CONVEX_HULL", "MESH"},
					},
				},
				"required": []string{"object_names"},
			},
		},
		{
			Name:        "setup_fluid_domain",
			Description: "Create a Mantaflow fluid domain (liquid or gas).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"domain_name": map[string]any{
						"type":        "string",
						"description": "Domain object name (default FluidDomain)",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "[x, y, z] location",
						"items":       map[string]any{"type": "number"},
					},
					"size": map[string]any{
						"type":        "number",
						"description": "Domain cube size (default 4.0)",
					},
					"resolution": map[string]any{
						"type":        "integer",
						"description": "Max resolution (default 64)",
					},
					"cache_dir": map[string]any{
						"type":        "string",
						"description": "Cache directory (default //fluid_cache)",
					},
					"domain_type": map[string]any{
						"type":        "string",
						"description": "LIQUID or GAS (default LIQUID)",
						"enum":        []string{"LIQUID", "GAS"},
					},
				},
			},
		},
		{
			Name:        "setup_fluid_inflow",
			Description: "Add a fluid inflow (flow) object to a simulation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object name to configure as inflow",
					},
					"flow_type": map[string]any{
						"type":        "string",
						"description": "LIQUID or GAS (default LIQUID)",
						"enum":        []string{"LIQUID", "GAS"},
					},
					"flow_behavior": map[string]any{
						"type":        "string",
						"description": "INFLOW, OUTFLOW, or GEOMETRY (default INFLOW)",
						"enum":        []string{"INFLOW", "OUTFLOW", "GEOMETRY"},
					},
					"use_particle_size": map[string]any{
						"type":        "number",
						"description": "Particle sampling size (default 0.1)",
					},
				},
				"required": []string{"object_name"},
			},
		},
		{
			Name:        "setup_effector",
			Description: "Set up collision/effector objects for fluid or physics simulations.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_names": map[string]any{
						"type":        "array",
						"description": "Object names to configure as effectors",
						"items":       map[string]any{"type": "string"},
					},
					"effector_type": map[string]any{
						"type":        "string",
						"description": "Effector type (default COLLISION)",
						"enum":        []string{"COLLISION"},
					},
					"surface_distance": map[string]any{
						"type":        "number",
						"description": "Surface thickness (default 0.01)",
					},
				},
				"required": []string{"object_names"},
			},
		},
		{
			Name:        "setup_camera",
			Description: "Create a camera and optionally position it, set focal length, and make it active.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Camera name (default Camera)",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "[x, y, z] location",
						"items":       map[string]any{"type": "number"},
					},
					"rotation": map[string]any{
						"type":        "array",
						"description": "[rx, ry, rz] Euler rotation in radians",
						"items":       map[string]any{"type": "number"},
					},
					"focal_length": map[string]any{
						"type":        "number",
						"description": "Focal length in mm (default 50)",
					},
					"set_active": map[string]any{
						"type":        "boolean",
						"description": "Make this the scene active camera (default true)",
					},
					"use_existing": map[string]any{
						"type":        "string",
						"description": "Name of existing camera to configure instead of creating new",
					},
				},
			},
		},
		{
			Name:        "manage_collections",
			Description: "Create collections and move objects between them.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"collections": map[string]any{
						"type":        "array",
						"description": "Collection specs: {name, objects, parent, color_tag}",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":      map[string]any{"type": "string", "description": "Collection name"},
								"objects":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Object names to move into this collection"},
								"parent":    map[string]any{"type": "string", "description": "Parent collection name"},
								"color_tag": map[string]any{"type": "string", "description": "Color tag e.g. COLOR_01"},
							},
							"required": []string{"name"},
						},
					},
				},
				"required": []string{"collections"},
			},
		},
		{
			Name:        "set_render_engine",
			Description: "Switch the render engine (BLENDER_EEVEE, BLENDER_CYCLES, HYDRA_STORM).",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"engine": map[string]any{
						"type":        "string",
						"description": "Render engine identifier",
						"enum":        []string{"BLENDER_EEVEE", "BLENDER_EEVEE_NEXT", "BLENDER_CYCLES", "HYDRA_STORM"},
					},
				},
				"required": []string{"engine"},
			},
		},
		{
			Name:        "set_render_format",
			Description: "Set render output format, resolution, color mode, and color depth.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"file_format": map[string]any{
						"type":        "string",
						"description": "Output format",
						"enum":        []string{"PNG", "JPEG", "BMP", "TARGA", "TARGA_RAW", "TIFF", "OPEN_EXR", "OPEN_EXR_MULTILAYER", "HDR", "AVIF", "WEBP", "CINEON", "DPX", "IRIS", "JPEG2000", "FFMPEG"},
					},
					"color_mode": map[string]any{
						"type":        "string",
						"description": "Color mode",
						"enum":        []string{"BW", "RGB", "RGBA"},
					},
					"color_depth": map[string]any{
						"type":        "string",
						"description": "Bit depth (format-dependent: 8, 10, 12, 16, 32)",
					},
					"resolution_x": map[string]any{
						"type":        "integer",
						"description": "Render width in pixels",
					},
					"resolution_y": map[string]any{
						"type":        "integer",
						"description": "Render height in pixels",
					},
					"resolution_percentage": map[string]any{
						"type":        "integer",
						"description": "Resolution percentage (1-100)",
					},
				},
			},
		},
		{
			Name:        "set_world_environment",
			Description: "Configure world environment: background color, strength, or environment texture.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"color": map[string]any{
						"type":        "array",
						"description": "Background RGB color [r, g, b], values 0-1",
						"items":       map[string]any{"type": "number"},
					},
					"strength": map[string]any{
						"type":        "number",
						"description": "Background strength (default 1.0)",
					},
					"texture_path": map[string]any{
						"type":        "string",
						"description": "Path to environment texture image (HDRI etc.)",
					},
				},
			},
		},
	}
}

// RegisterTools creates a ToolRegistry with all Blender tools.
func RegisterTools(bc *blender.Client) *ToolRegistry {
	tr := NewToolRegistry()

	allTools := []struct {
		tool    Tool
		handler ToolHandler
	}{
		{ToolDefs()[0], handleExecuteCode(bc)},
		{ToolDefs()[1], handleGetSceneInfo(bc)},
		{ToolDefs()[2], handleListObjects(bc)},
		{ToolDefs()[3], handleGetObjectInfo(bc)},
		{ToolDefs()[4], handleCreateMeshObject(bc)},
		{ToolDefs()[5], handleDeleteObject(bc)},
		{ToolDefs()[6], handleSetObjectLocation(bc)},
		{ToolDefs()[7], handleSetObjectMaterial(bc)},
		{ToolDefs()[8], handleRenderScene(bc)},
		{ToolDefs()[9], handleGetViewportScreenshot(bc)},
		{ToolDefs()[10], handleImport3DFile(bc)},
		{ToolDefs()[11], handleExport3DFile(bc)},
		{ToolDefs()[12], handleSetViewportCamera(bc)},
		{ToolDefs()[13], handleAddLight(bc)},
		{ToolDefs()[14], handleAddCamera(bc)},
		{ToolDefs()[15], handleSetObjectRotation(bc)},
		{ToolDefs()[16], handleSetObjectScale(bc)},
		{ToolDefs()[17], handleDuplicateObject(bc)},
		{ToolDefs()[18], handleUndo(bc)},
		{ToolDefs()[19], handleRedo(bc)},
		{ToolDefs()[20], handleListMaterials(bc)},
		{ToolDefs()[21], handleListCollections(bc)},
		{ToolDefs()[22], handleRenderAnimation(bc)},
		{ToolDefs()[23], handleGetNodeTree(bc)},
		{ToolDefs()[24], handleGetAnimationData(bc)},
		{ToolDefs()[25], handleGetObjectHierarchy(bc)},
		{ToolDefs()[26], handleManageModifiers(bc)},
		{ToolDefs()[27], handleManageConstraints(bc)},
		{ToolDefs()[28], handleManagePhysics(bc)},
		{ToolDefs()[29], handleSetMaterialColor(bc)},
		{ToolDefs()[30], handleSetMaterialTexture(bc)},
		{ToolDefs()[31], handleGetTransform(bc)},
		{ToolDefs()[32], handleMaterialCreate(bc)},
		{ToolDefs()[33], handleMaterialAssign(bc)},
		{ToolDefs()[34], handleSaveBlend(bc)},
		{ToolDefs()[35], handleApplyTransforms(bc)},
		{ToolDefs()[36], handleSetFrameRange(bc)},
		{ToolDefs()[37], handleSetViewportShading(bc)},
		{ToolDefs()[38], handleExportGLTF(bc)},
		{ToolDefs()[39], handleExportOBJ(bc)},
		{ToolDefs()[40], handleExportFBX(bc)},
		{ToolDefs()[41], handleSetupKeyframes(bc)},
		{ToolDefs()[42], handleSetupRigidBody(bc)},
		{ToolDefs()[43], handleSetupFluidDomain(bc)},
		{ToolDefs()[44], handleSetupFluidInflow(bc)},
		{ToolDefs()[45], handleSetupEffector(bc)},
		{ToolDefs()[46], handleSetupCamera(bc)},
		{ToolDefs()[47], handleManageCollections(bc)},
	}

	for _, entry := range allTools {
		tr.Register(entry.tool, entry.handler)
	}

	return tr
}

// helper to get a string param
func getStringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// helper to get a []float64 param
func getFloatSliceArg(args map[string]any, key string) []float64 {
	if v, ok := args[key]; ok {
		if arr, ok := v.([]any); ok {
			result := make([]float64, len(arr))
			for i, item := range arr {
				if f, ok := item.(float64); ok {
					result[i] = f
				}
			}
			return result
		}
	}
	return nil
}

func fmtFloats(vals []float64) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return strings.Join(parts, ", ")
}

// --- Tool handlers ---

func handleExecuteCode(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := getStringArg(args, "code")
		if code == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "code is required"}
		}
		result, err := bc.ExecuteCode(code, true)
		if err != nil {
			return nil, &Error{Code: CodeInternalError, Message: err.Error()}
		}
		text := formatExecResult(result)
		return ToolCallResult{
			Content: []ContentBlock{NewTextContent(text)},
			IsError: result.Status == "error",
		}, nil
	}
}

func handleGetSceneInfo(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy, json
scene = bpy.context.scene
cam = scene.camera
result = {
    "scene_name": scene.name,
    "camera": cam.name if cam else None,
    "render_engine": scene.render.engine,
    "frame_current": scene.frame_current,
    "frame_start": scene.frame_start,
    "frame_end": scene.frame_end,
    "fps": scene.render.fps,
    "resolution_x": scene.render.resolution_x,
    "resolution_y": scene.render.resolution_y,
    "num_objects": len(scene.objects),
    "view_layers": [vl.name for vl in scene.view_layers],
}`
		return execAndFormat(bc, code)
	}
}

func handleListObjects(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filterType := getStringArg(args, "filter_type")

		filterClause := ""
		if filterType != "" {
			filterClause = fmt.Sprintf(` if obj.type == "%s"`, filterType)
		}

		code := fmt.Sprintf(`import bpy, json
scene = bpy.context.scene
objects = []
for obj in scene.objects%s:
    objects.append({
        "name": obj.name,
        "type": obj.type,
        "location": [round(v, 4) for v in obj.location],
        "visible_get": obj.visible_get(),
    })
result = {"objects": objects, "count": len(objects)}
`, filterClause)
		return execAndFormat(bc, code)
	}
}

func handleGetObjectInfo(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		if name == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "name is required"}
		}

		code := fmt.Sprintf(`import bpy, json
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found: %s"}
else:
    result = {
        "name": obj.name,
        "type": obj.type,
        "location": [round(v, 4) for v in obj.location],
        "rotation_euler": [round(v, 4) for v in obj.rotation_euler],
        "scale": [round(v, 4) for v in obj.scale],
        "visible_get": obj.visible_get(),
        "hide_viewport": obj.hide_viewport,
        "hide_render": obj.hide_render,
        "materials": [m.name if m else None for m in obj.data.materials] if hasattr(obj, 'data') and hasattr(obj.data, 'materials') else [],
        "data_type": type(obj.data).__name__ if obj.data else None,
    }
`, name, name)
		return execAndFormat(bc, code)
	}
}

func handleCreateMeshObject(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		primitive := getStringArg(args, "primitive")
		if primitive == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "primitive is required"}
		}

		loc := getFloatSliceArg(args, "location")
		locArg := "0,0,0"
		if len(loc) == 3 {
			locArg = fmtFloats(loc)
		}

		name := getStringArg(args, "name")
		nameArg := ""
		if name != "" {
			nameArg = fmt.Sprintf(`    obj.name = "%s"
`, name)
		}

		size := 2.0
		if v, ok := args["size"].(float64); ok && v > 0 {
			size = v
		}

		// Data API mesh creation — more stable than bpy.ops for live bridge sessions.
		// Falls back to bpy.ops for monkey (Suzanne) which has no data API path.
		var code string

		switch primitive {
		case "cube":
			code = fmt.Sprintf(`import bpy, bmesh, mathutils
h = %g / 2.0
mesh = bpy.data.meshes.new("CubeMesh")
bm = bmesh.new()
bmesh.ops.create_cube(bm, size=%g)
bm.to_mesh(mesh)
bm.free()
mesh.update()
obj = bpy.data.objects.new("Cube", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, size, locArg, nameArg)

		case "sphere":
			code = fmt.Sprintf(`import bpy, math
segments = 32
rings = 16
radius = %g / 2.0
verts = []
faces = []
for i in range(rings + 1):
    phi = math.pi * i / rings
    for j in range(segments):
        theta = 2.0 * math.pi * j / segments
        x = radius * math.sin(phi) * math.cos(theta)
        y = radius * math.sin(phi) * math.sin(theta)
        z = radius * math.cos(phi)
        verts.append((x, y, z))
for i in range(rings):
    for j in range(segments):
        a = i * segments + j
        b = i * segments + (j + 1) %% segments
        c = (i + 1) * segments + (j + 1) %% segments
        d = (i + 1) * segments + j
        faces.append((a, b, c, d))
mesh = bpy.data.meshes.new("SphereMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("UV Sphere", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, locArg, nameArg)

		case "plane":
			code = fmt.Sprintf(`import bpy
h = %g / 2.0
verts = [(-h, -h, 0), (h, -h, 0), (h, h, 0), (-h, h, 0)]
faces = [(0, 1, 2, 3)]
mesh = bpy.data.meshes.new("PlaneMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("Plane", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, locArg, nameArg)

		case "cylinder":
			code = fmt.Sprintf(`import bpy, math
segments = 32
radius = %g / 2.0
depth = %g
verts = []
faces = []
for i in range(segments):
    angle = 2.0 * math.pi * i / segments
    x = radius * math.cos(angle)
    y = radius * math.sin(angle)
    verts.append((x, y, -depth / 2.0))
    verts.append((x, y, depth / 2.0))
for i in range(segments):
    n = (i + 1) %% segments
    faces.append((i * 2, n * 2, n * 2 + 1, i * 2 + 1))
top_center = len(verts)
verts.append((0, 0, depth / 2.0))
bottom_center = len(verts)
verts.append((0, 0, -depth / 2.0))
top_ring = [(i * 2 + 1) for i in range(segments)]
bottom_ring = [(i * 2) for i in range(segments)]
faces.append(tuple(top_ring))
faces.append(tuple(reversed(bottom_ring)))
mesh = bpy.data.meshes.new("CylinderMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("Cylinder", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, size, locArg, nameArg)

		case "cone":
			code = fmt.Sprintf(`import bpy, math
segments = 32
radius = %g / 2.0
depth = %g
verts = []
faces = []
for i in range(segments):
    angle = 2.0 * math.pi * i / segments
    x = radius * math.cos(angle)
    y = radius * math.sin(angle)
    verts.append((x, y, -depth / 2.0))
apex = len(verts)
verts.append((0, 0, depth / 2.0))
for i in range(segments):
    n = (i + 1) %% segments
    faces.append((i, n, apex))
base_ring = list(range(segments))
faces.append(tuple(reversed(base_ring)))
mesh = bpy.data.meshes.new("ConeMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("Cone", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, size, locArg, nameArg)

		case "torus":
			code = fmt.Sprintf(`import bpy, math
major = %g / 2.0
minor = major / 4.0
major_segments = 48
minor_segments = 12
verts = []
faces = []
for i in range(major_segments):
    theta = 2.0 * math.pi * i / major_segments
    ct, st = math.cos(theta), math.sin(theta)
    for j in range(minor_segments):
        phi = 2.0 * math.pi * j / minor_segments
        cp, sp = math.cos(phi), math.sin(phi)
        x = (major + minor * cp) * ct
        y = (major + minor * cp) * st
        z = minor * sp
        verts.append((x, y, z))
for i in range(major_segments):
    ni = (i + 1) %% major_segments
    for j in range(minor_segments):
        nj = (j + 1) %% minor_segments
        a = i * minor_segments + j
        b = ni * minor_segments + j
        c = ni * minor_segments + nj
        d = i * minor_segments + nj
        faces.append((a, b, c, d))
mesh = bpy.data.meshes.new("TorusMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("Torus", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
%sobj.select_set(True)
bpy.context.view_layer.objects.active = obj
result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
`, size, locArg, nameArg)

		case "monkey":
			// No data API path for Suzanne — use bpy.ops as fallback
			locParam := ""
			if len(loc) == 3 {
				locParam = fmt.Sprintf(", location=(%s)", fmtFloats(loc))
			}
			code = fmt.Sprintf(`import bpy
bpy.ops.object.select_all(action='SELECT')
bpy.ops.object.delete()
bpy.ops.mesh.primitive_monkey_add(%s)
obj = bpy.context.active_object
if obj:
%s    result = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
else:
    result = {"error": "failed to create object"}
`, locParam, nameArg)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown primitive: " + primitive}
		}

		return execAndFormat(bc, code)
	}
}

func handleDeleteObject(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		if name == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "name is required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    bpy.data.objects.remove(obj, do_unlink=True)
    result = {"deleted": "%s"}
`, name, name)
		return execAndFormat(bc, code)
	}
}

func handleSetObjectLocation(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		loc := getFloatSliceArg(args, "location")
		if name == "" || len(loc) != 3 {
			return nil, &Error{Code: CodeInvalidParams, Message: "name and location[3] are required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    obj.location = (%s)
    result = {"name": obj.name, "location": [round(v, 4) for v in obj.location]}
`, name, fmtFloats(loc))
		return execAndFormat(bc, code)
	}
}

func handleSetObjectMaterial(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		matName := getStringArg(args, "material_name")
		if name == "" || matName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "name and material_name are required"}
		}

		color := getFloatSliceArg(args, "color")
		colorArg := ""
		if len(color) >= 3 {
			a := 1.0
			if len(color) >= 4 {
				a = color[3]
			}
			colorArg = fmt.Sprintf(`
mat.use_nodes = True
bsdf = mat.node_tree.nodes.get("Principled BSDF")
if bsdf:
    bsdf.inputs["Base Color"].default_value = (%g, %g, %g, %g)`, color[0], color[1], color[2], a)
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mat = bpy.data.materials.get("%s")
    if mat is None:
        mat = bpy.data.materials.new(name="%s")
%s
    if len(obj.data.materials) == 0:
        obj.data.materials.append(mat)
    else:
        obj.data.materials[0] = mat
    result = {"name": obj.name, "material": mat.name}
`, name, matName, matName, colorArg)
		return execAndFormat(bc, code)
	}
}

func handleRenderScene(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		outputPath := getStringArg(args, "output_path")
		if outputPath == "" {
			outputPath = "/tmp/blender_render.png"
		}

		resXArg := ""
		if v, ok := args["resolution_x"].(float64); ok && v > 0 {
			resXArg = fmt.Sprintf("\nscene.render.resolution_x = %d", int(v))
		}
		resYArg := ""
		if v, ok := args["resolution_y"].(float64); ok && v > 0 {
			resYArg = fmt.Sprintf("\nscene.render.resolution_y = %d", int(v))
		}

		code := fmt.Sprintf(`import bpy, os
scene = bpy.context.scene
scene.render.filepath = "%s"%s%s
bpy.ops.render.render(write_still=True)
result = {"rendered": True, "output_path": "%s"}
`, outputPath, resXArg, resYArg, outputPath)
		return execAndFormat(bc, code)
	}
}

func handleGetViewportScreenshot(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy, base64, tempfile, os
path = os.path.join(tempfile.gettempdir(), "blender_viewport.png")
bpy.ops.screen.screenshot(filepath=path)
with open(path, "rb") as f:
    data = base64.b64encode(f.read()).decode("ascii")
result = {"image_base64": data, "path": path}`
		return execAndFormat(bc, code)
	}
}

func handleImport3DFile(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")
		if filepath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "filepath is required"}
		}

		ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])

		var importFn string
		switch ext {
		case ".glb", ".gltf":
			importFn = fmt.Sprintf(`bpy.ops.import_scene.gltf(filepath="%s")`, filepath)
		case ".fbx":
			importFn = fmt.Sprintf(`bpy.ops.import_scene.fbx(filepath="%s")`, filepath)
		case ".obj":
			importFn = fmt.Sprintf(`bpy.ops.wm.obj_import(filepath="%s")`, filepath)
		case ".stl":
			importFn = fmt.Sprintf(`bpy.ops.wm.stl_import(filepath="%s")`, filepath)
		case ".ply":
			importFn = fmt.Sprintf(`bpy.ops.wm.ply_import(filepath="%s")`, filepath)
		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unsupported file type: " + ext}
		}

		code := fmt.Sprintf(`import bpy
selected = bpy.context.selected_objects[:]
%s
new_objects = [o for o in bpy.context.selected_objects if o not in selected]
result = {"imported": [o.name for o in new_objects], "count": len(new_objects)}
`, importFn)
		return execAndFormat(bc, code)
	}
}

func handleExport3DFile(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")
		if filepath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "filepath is required"}
		}

		ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])

		selArg := ""
		if v, ok := args["use_selection"].(bool); ok && v {
			selArg = ", use_selection=True"
		}

		var exportFn string
		switch ext {
		case ".glb", ".gltf":
			exportFn = fmt.Sprintf(`bpy.ops.export_scene.gltf(filepath="%s"%s)`, filepath, selArg)
		case ".fbx":
			exportFn = fmt.Sprintf(`bpy.ops.export_scene.fbx(filepath="%s"%s)`, filepath, selArg)
		case ".obj":
			exportFn = fmt.Sprintf(`bpy.ops.wm.obj_export(filepath="%s"%s)`, filepath, selArg)
		case ".stl":
			exportFn = fmt.Sprintf(`bpy.ops.wm.stl_export(filepath="%s"%s)`, filepath, selArg)
		case ".ply":
			exportFn = fmt.Sprintf(`bpy.ops.wm.ply_export(filepath="%s"%s)`, filepath, selArg)
		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unsupported file type: " + ext}
		}

		code := fmt.Sprintf(`import bpy
%s
result = {"exported": "%s"}
`, exportFn, filepath)
		return execAndFormat(bc, code)
	}
}

func handleSetViewportCamera(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		loc := getFloatSliceArg(args, "location")
		rot := getFloatSliceArg(args, "rotation")
		if len(loc) != 3 || len(rot) != 3 {
			return nil, &Error{Code: CodeInvalidParams, Message: "location and rotation must be [x, y, z]"}
		}

		code := fmt.Sprintf(`import bpy, math
area = None
for a in bpy.context.screen.areas:
    if a.type == 'VIEW_3D':
        area = a
        break
if area:
    rv3d = area.spaces.active.region_3d
    rv3d.view_location = (%s)
    rv3d.view_rotation = (0.5, -0.5, 0.5, 0.5)
    from mathutils import Euler
    rot = Euler((%s), 'XYZ')
    rv3d.view_rotation = rot.to_quaternion()
    result = {"set": True}
else:
    result = {"error": "no 3D viewport found"}
`, fmtFloats(loc), fmtFloats(rot))
		return execAndFormat(bc, code)
	}
}

func handleAddLight(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		lightType := getStringArg(args, "light_type")
		if lightType == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "light_type is required"}
		}

		loc := getFloatSliceArg(args, "location")
		locArg := ""
		if len(loc) == 3 {
			locArg = fmt.Sprintf(", location=(%s)", fmtFloats(loc))
		}
		energyArg := ""
		if v, ok := args["energy"].(float64); ok && v > 0 {
			energyArg = fmt.Sprintf("\nlight.data.energy = %g", v)
		}
		color := getFloatSliceArg(args, "color")
		colorArg := ""
		if len(color) >= 3 {
			colorArg = fmt.Sprintf("\nlight.data.color = (%s)", fmtFloats(color[:3]))
		}
		name := getStringArg(args, "name")
		nameArg := ""
		if name != "" {
			nameArg = fmt.Sprintf("\nlight.name = \"%s\"", name)
		}

		code := fmt.Sprintf(`import bpy
bpy.ops.object.light_add(type="%s"%s)
light = bpy.context.active_object
if light:%s%s%s
result = {"name": light.name, "type": light.data.type} if light else {"error": "failed to add light"}
`, lightType, locArg, nameArg, energyArg, colorArg)
		return execAndFormat(bc, code)
	}
}

func handleAddCamera(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		loc := getFloatSliceArg(args, "location")
		locArg := ""
		if len(loc) == 3 {
			locArg = fmt.Sprintf(", location=(%s)", fmtFloats(loc))
		}
		rot := getFloatSliceArg(args, "rotation")
		rotArg := ""
		if len(rot) == 3 {
			rotArg = fmt.Sprintf("\ncam.rotation_euler = (%s)", fmtFloats(rot))
		}
		name := getStringArg(args, "name")
		nameArg := ""
		if name != "" {
			nameArg = fmt.Sprintf("\ncam.name = \"%s\"", name)
		}

		code := fmt.Sprintf(`import bpy
bpy.ops.object.camera_add(%s)
cam = bpy.context.active_object
if cam:%s%s
result = {"name": cam.name, "location": [round(v, 4) for v in cam.location]} if cam else {"error": "failed to add camera"}
`, locArg, nameArg, rotArg)
		return execAndFormat(bc, code)
	}
}

func handleSetObjectRotation(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		rot := getFloatSliceArg(args, "rotation")
		if name == "" || len(rot) != 3 {
			return nil, &Error{Code: CodeInvalidParams, Message: "name and rotation[3] are required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    obj.rotation_euler = (%s)
    result = {"name": obj.name, "rotation_euler": [round(v, 4) for v in obj.rotation_euler]}
`, name, fmtFloats(rot))
		return execAndFormat(bc, code)
	}
}

func handleSetObjectScale(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		scale := getFloatSliceArg(args, "scale")
		if name == "" || len(scale) != 3 {
			return nil, &Error{Code: CodeInvalidParams, Message: "name and scale[3] are required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    obj.scale = (%s)
    result = {"name": obj.name, "scale": [round(v, 4) for v in obj.scale]}
`, name, fmtFloats(scale))
		return execAndFormat(bc, code)
	}
}

func handleDuplicateObject(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		if name == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "name is required"}
		}

		newName := getStringArg(args, "new_name")
		newNameArg := ""
		if newName != "" {
			newNameArg = fmt.Sprintf("\nnew_obj.name = \"%s\"", newName)
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    new_obj = obj.copy()
    if obj.data:
        new_obj.data = obj.data.copy()
    bpy.context.collection.objects.link(new_obj)
%s
    result = {"name": new_obj.name, "type": new_obj.type, "location": [round(v, 4) for v in new_obj.location]}
`, name, newNameArg)
		return execAndFormat(bc, code)
	}
}

func handleUndo(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy
bpy.ops.ed.undo()
result = {"undone": true}`
		return execAndFormat(bc, code)
	}
}

func handleRedo(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy
bpy.ops.ed.redo()
result = {"redone": true}`
		return execAndFormat(bc, code)
	}
}

func handleListMaterials(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy
materials = []
for mat in bpy.data.materials:
    materials.append({
        "name": mat.name,
        "use_nodes": mat.use_nodes,
        "users": mat.users,
    })
result = {"materials": materials, "count": len(materials)}`
		return execAndFormat(bc, code)
	}
}

func handleListCollections(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy
collections = []
for col in bpy.data.collections:
    collections.append({
        "name": col.name,
        "object_count": len(col.objects),
        "parent": col.parent.name if col.parent else None,
    })
result = {"collections": collections, "count": len(collections)}`
		return execAndFormat(bc, code)
	}
}

func handleRenderAnimation(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		outputDir := getStringArg(args, "output_dir")
		if outputDir == "" {
			outputDir = "/tmp/blender_anim"
		}

		startArg := ""
		if v, ok := args["frame_start"].(float64); ok && v >= 0 {
			startArg = fmt.Sprintf("\nscene.frame_start = %d", int(v))
		}
		endArg := ""
		if v, ok := args["frame_end"].(float64); ok && v >= 0 {
			endArg = fmt.Sprintf("\nscene.frame_end = %d", int(v))
		}

		code := fmt.Sprintf(`import bpy, os
scene = bpy.context.scene
os.makedirs("%s", exist_ok=True)
scene.render.filepath = os.path.join("%s", "frame_####")
%s%s
bpy.ops.render.render(animation=True)
result = {"rendered": True, "output_dir": "%s"}
`, outputDir, outputDir, startArg, endArg, outputDir)
		return execAndFormat(bc, code)
	}
}

func handleSetMaterialTexture(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		matName := getStringArg(args, "material_name")
		imagePath := getStringArg(args, "image_path")
		if matName == "" || imagePath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "material_name and image_path are required"}
		}

		code := fmt.Sprintf(`import bpy, os
mat = bpy.data.materials.get("%s")
if mat is None:
    result = {"error": "material not found"}
else:
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes.get("Principled BSDF")
    if bsdf is None:
        result = {"error": "no Principled BSDF node found"}
    else:
        img_node = mat.node_tree.nodes.new("ShaderNodeTexImage")
        img_node.image = bpy.data.images.load("%s")
        mat.node_tree.links.new(img_node.outputs["Color"], bsdf.inputs["Base Color"])
        result = {"material": mat.name, "texture": "%s"}
`, matName, imagePath, imagePath)
		return execAndFormat(bc, code)
	}
}

func handleSetMaterialColor(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		matName := getStringArg(args, "material_name")
		color := getFloatSliceArg(args, "color")
		if matName == "" || len(color) < 3 {
			return nil, &Error{Code: CodeInvalidParams, Message: "material_name and color[3+] are required"}
		}

		a := 1.0
		if len(color) >= 4 {
			a = color[3]
		}

		code := fmt.Sprintf(`import bpy
mat = bpy.data.materials.get("%s")
if mat is None:
    result = {"error": "material not found"}
else:
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes.get("Principled BSDF")
    if bsdf:
        bsdf.inputs["Base Color"].default_value = (%g, %g, %g, %g)
        result = {"material": mat.name, "color": [%g, %g, %g, %g]}
    else:
        result = {"error": "no Principled BSDF node found"}
`, matName, color[0], color[1], color[2], a, color[0], color[1], color[2], a)
		return execAndFormat(bc, code)
	}
}

func handleGetNodeTree(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		slotIdx := 0
		if v, ok := args["material_slot"].(float64); ok && v >= 0 {
			slotIdx = int(v)
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
elif not obj.data or not hasattr(obj.data, "materials") or len(obj.data.materials) == 0:
    result = {"error": "object has no materials"}
elif %d >= len(obj.data.materials):
    result = {"error": "material slot index out of range"}
else:
    mat = obj.data.materials[%d]
    if mat is None or mat.node_tree is None:
        result = {"error": "material has no node tree"}
    else:
        nodes = []
        for n in mat.node_tree.nodes:
            inputs = [{"name": inp.name, "type": inp.type, "default_value": list(inp.default_value) if inp.default_value is not None else None} for inp in n.inputs]
            outputs = [{"name": out.name, "type": out.type, "default_value": list(out.default_value) if out.default_value is not None else None} for out in n.outputs]
            nodes.append({
                "name": n.name,
                "type": n.type,
                "location": [round(v, 4) for v in n.location],
                "inputs": inputs,
                "outputs": outputs,
            })
        links = []
        for l in mat.node_tree.links:
            links.append({
                "from_node": l.from_node.name,
                "from_socket": l.from_socket.name,
                "to_node": l.to_node.name,
                "to_socket": l.to_socket.name,
            })
        result = {"material": mat.name, "node_count": len(nodes), "link_count": len(links), "nodes": nodes, "links": links}
`, objName, slotIdx, slotIdx)
		return execAndFormat(bc, code)
	}
}

func handleGetAnimationData(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		objArg := ""
		if objName != "" {
			objArg = fmt.Sprintf(`obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
    return
`, objName)
		} else {
			objArg = `obj = bpy.context.active_object
if obj is None:
    result = {"error": "no active object"}
    return
`
		}

		code := fmt.Sprintf(`import bpy
%s
keyframes = {}
if obj.animation_data and obj.animation_data.action:
    for fc in obj.animation_data.action.fcurves:
        kfs = [{"frame": kp.co[0], "value": kp.co[1]} for kp in fc.keyframe_points]
        keyframes[fc.data_path + ("[" + str(fc.array_index) + "]" if fc.array_index >= 0 else "")] = kfs

nla_strips = []
if obj.animation_data and obj.animation_data.nla_tracks:
    for track in obj.animation_data.nla_tracks:
        for strip in track.strips:
            nla_strips.append({
                "track": track.name,
                "name": strip.name,
                "type": strip.type,
                "frame_start": strip.frame_start,
                "frame_end": strip.frame_end,
                "action": strip.action.name if strip.action else None,
            })

drivers = []
if obj.animation_data and obj.animation_data.drivers:
    for dr in obj.animation_data.drivers:
        drivers.append({
            "data_path": dr.data_path,
            "array_index": dr.array_index,
            "expression": dr.driver.expression if dr.driver else None,
        })

result = {
    "object": obj.name,
    "has_animation_data": obj.animation_data is not None,
    "keyframe_channels": len(keyframes),
    "keyframes": keyframes,
    "nla_strips": nla_strips,
    "driver_count": len(drivers),
    "drivers": drivers,
}
`, objArg)
		return execAndFormat(bc, code)
	}
}

func handleGetObjectHierarchy(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		maxDepth := 3
		if v, ok := args["max_depth"].(float64); ok && v > 0 {
			maxDepth = int(v)
		}

		rootRef := "bpy.context.active_object"
		if objName != "" {
			rootRef = fmt.Sprintf(`bpy.data.objects.get("%s")`, objName)
		}

		code := fmt.Sprintf(`import bpy
def build_tree(obj, depth, max_d):
    if obj is None or depth > max_d:
        return None
    node = {"name": obj.name, "type": obj.type, "children": []}
    for child in obj.children:
        c = build_tree(child, depth + 1, max_d)
        if c:
            node["children"].append(c)
    return node

root = %s
if root is None:
    result = {"error": "object not found or no active object"}
else:
    tree = build_tree(root, 0, %d)
    result = {"hierarchy": tree}
`, rootRef, maxDepth)
		return execAndFormat(bc, code)
	}
}

func handleManageModifiers(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		action := getStringArg(args, "action")
		if action == "" {
			action = "list"
		}

		switch action {
		case "list":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mods = [{"name": m.name, "type": m.type, "show_viewport": m.show_viewport, "show_render": m.show_render} for m in obj.modifiers]
    result = {"object": obj.name, "modifiers": mods, "count": len(mods)}
`, objName)
			return execAndFormat(bc, code)

		case "add":
			modType := getStringArg(args, "modifier_type")
			if modType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "modifier_type is required for add action"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    try:
        mod = obj.modifiers.new(name="%s", type="%s")
        result = {"added": mod.name, "type": mod.type, "object": obj.name}
    except Exception as e:
        result = {"error": str(e)}
`, objName, modType, modType)
			return execAndFormat(bc, code)

		case "remove":
			modName := getStringArg(args, "modifier_name")
			if modName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "modifier_name is required for remove action"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.get("%s")
    if mod is None:
        result = {"error": "modifier not found"}
    else:
        obj.modifiers.remove(mod)
        result = {"removed": "%s", "object": obj.name}
`, objName, modName, modName)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleManageConstraints(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		action := getStringArg(args, "action")
		if action == "" {
			action = "list"
		}

		switch action {
		case "list":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    cons = [{"name": c.name, "type": c.type, "target": c.target.name if c.target else None, "enabled": not c.mute} for c in obj.constraints]
    result = {"object": obj.name, "constraints": cons, "count": len(cons)}
`, objName)
			return execAndFormat(bc, code)

		case "add":
			conType := getStringArg(args, "constraint_type")
			if conType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "constraint_type is required for add action"}
			}
			targetName := getStringArg(args, "target_name")
			targetArg := ""
			if targetName != "" {
				targetArg = fmt.Sprintf(`
c.target = bpy.data.objects.get("%s")`, targetName)
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    try:
        c = obj.constraints.new(type="%s")
%s
        result = {"added": c.name, "type": c.type, "object": obj.name}
    except Exception as e:
        result = {"error": str(e)}
`, objName, conType, targetArg)
			return execAndFormat(bc, code)

		case "remove":
			conName := getStringArg(args, "constraint_name")
			if conName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "constraint_name is required for remove action"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    c = None
    for con in obj.constraints:
        if con.name == "%s":
            c = con
            break
    if c is None:
        result = {"error": "constraint not found"}
    else:
        obj.constraints.remove(c)
        result = {"removed": "%s", "object": obj.name}
`, objName, conName, conName)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleManagePhysics(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		action := getStringArg(args, "action")
		if action == "" {
			action = "list"
		}

		switch action {
		case "list":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    phys = []
    if obj.rigid_body:
        phys.append({"type": "RIGID_BODY", "mass": obj.rigid_body.mass, "kinematic": obj.rigid_body.kinematic, "collision_shape": obj.rigid_body.collision_shape})
    if obj.modifiers.get("Cloth"):
        phys.append({"type": "CLOTH"})
    if obj.modifiers.get("Fluid"):
        phys.append({"type": "FLUID"})
    if obj.field:
        phys.append({"type": "FORCE_FIELD", "field_type": obj.field.type})
    if obj.modifiers.get("Soft Body"):
        phys.append({"type": "SOFT_BODY"})
    if obj.modifiers.get("Particle System"):
        phys.append({"type": "PARTICLE_SYSTEM"})
    result = {"object": obj.name, "physics": phys, "count": len(phys)}
`, objName)
			return execAndFormat(bc, code)

		case "add":
			physType := getStringArg(args, "physics_type")
			if physType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "physics_type is required for add action"}
			}

			var code string
			switch physType {
			case "RIGID_BODY":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.rigidbody.object_add()
    result = {"added": "RIGID_BODY", "object": obj.name}
`, objName)
			case "CLOTH":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Cloth", type="CLOTH")
    result = {"added": "CLOTH", "object": obj.name}
`, objName)
			case "FLUID":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Fluid", type="FLUID")
    mod.fluid_type = "DOMAIN"
    result = {"added": "FLUID", "object": obj.name}
`, objName)
			case "FORCE_FIELD":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    obj.field.type = "WIND"
    result = {"added": "FORCE_FIELD", "object": obj.name}
`, objName)
			case "SOFT_BODY":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Soft Body", type="SOFT_BODY")
    result = {"added": "SOFT_BODY", "object": obj.name}
`, objName)
			case "PARTICLE_SYSTEM":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Particle", type="PARTICLE_SYSTEM")
    result = {"added": "PARTICLE_SYSTEM", "object": obj.name}
`, objName)
			default:
				return nil, &Error{Code: CodeInvalidParams, Message: "unknown physics type: " + physType}
			}
			return execAndFormat(bc, code)

		case "remove":
			physType := getStringArg(args, "physics_type")
			if physType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "physics_type is required for remove action"}
			}

			var code string
			switch physType {
			case "RIGID_BODY":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.rigidbody.object_remove()
    result = {"removed": "RIGID_BODY", "object": obj.name}
`, objName)
			default:
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.get("%s")
    if mod is None:
        mod = obj.modifiers.get("%s")
    if mod is None:
        result = {"error": "physics modifier not found"}
    else:
        obj.modifiers.remove(mod)
        result = {"removed": "%s", "object": obj.name}
`, objName, physType, strings.Title(strings.ToLower(physType)), physType)
			}
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleGetTransform(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		name := getStringArg(args, "name")
		if name == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "name is required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    result = {
        "name": obj.name,
        "location": [round(v, 4) for v in obj.location],
        "rotation_euler": [round(v, 4) for v in obj.rotation_euler],
        "scale": [round(v, 4) for v in obj.scale],
    }
`, name)
		return execAndFormat(bc, code)
	}
}

func handleMaterialCreate(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		matName := getStringArg(args, "name")
		if matName == "" {
			matName = "Material"
		}

		color := getFloatSliceArg(args, "color")
		colorArg := ""
		if len(color) >= 3 {
			a := 1.0
			if len(color) >= 4 {
				a = color[3]
			}
			colorArg = fmt.Sprintf(`
bsdf = mat.node_tree.nodes.get("Principled BSDF")
if bsdf:
    bsdf.inputs["Base Color"].default_value = (%g, %g, %g, %g)`, color[0], color[1], color[2], a)
		}

		code := fmt.Sprintf(`import bpy
mat = bpy.data.materials.new(name="%s")
mat.use_nodes = True
%s
result = {"name": mat.name, "use_nodes": mat.use_nodes}
`, matName, colorArg)
		return execAndFormat(bc, code)
	}
}

func handleMaterialAssign(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		matName := getStringArg(args, "material_name")
		if objName == "" || matName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name and material_name are required"}
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
mat = bpy.data.materials.get("%s")
if obj is None:
    result = {"error": "object not found"}
elif mat is None:
    result = {"error": "material not found"}
else:
    if obj.data.materials:
        obj.data.materials[0] = mat
    else:
        obj.data.materials.append(mat)
    result = {"object": obj.name, "material": mat.name}
`, objName, matName)
		return execAndFormat(bc, code)
	}
}

func handleSaveBlend(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")

		if filepath == "" {
			code := `import bpy
if not bpy.data.filepath:
    result = {"error": "file has never been saved, provide a filepath"}
else:
    bpy.ops.wm.save_mainfile()
    result = {"filepath": bpy.data.filepath, "compressed": True}`
			return execAndFormat(bc, code)
		}

		code := fmt.Sprintf(`import bpy, os
abs_path = os.path.abspath("%s")
if not abs_path.endswith(".blend"):
    abs_path += ".blend"
bpy.ops.wm.save_as_mainfile(filepath=abs_path, compress=True, relative_remap=True)
result = {"filepath": abs_path, "compressed": True}
`, filepath)
		return execAndFormat(bc, code)
	}
}

func handleApplyTransforms(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		namesRaw, ok := args["object_names"].([]any)
		if !ok || len(namesRaw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_names is required and must be non-empty"}
		}

		names := make([]string, len(namesRaw))
		for i, v := range namesRaw {
			names[i] = fmt.Sprintf("%v", v)
		}

		locArg := "True"
		if v, ok := args["location"].(bool); ok && !v {
			locArg = "False"
		}
		rotArg := "True"
		if v, ok := args["rotation"].(bool); ok && !v {
			rotArg = "False"
		}
		sclArg := "True"
		if v, ok := args["scale"].(bool); ok && !v {
			sclArg = "False"
		}

		namesPy := strings.Join(func() []string {
			parts := make([]string, len(names))
			for i, n := range names {
				parts[i] = fmt.Sprintf(`"%s"`, n)
			}
			return parts
		}(), ", ")

		code := fmt.Sprintf(`import bpy
obj_names = [%s]
applied = []
skipped = []
bpy.ops.object.select_all(action="DESELECT")
for name in obj_names:
    obj = bpy.data.objects.get(name)
    if obj is None:
        skipped.append(name)
        continue
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.transform_apply(location=%s, rotation=%s, scale=%s)
    obj.select_set(False)
    applied.append(name)
result = {"applied": applied, "skipped": skipped}
`, namesPy, locArg, rotArg, sclArg)
		return execAndFormat(bc, code)
	}
}

func handleSetFrameRange(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		startArg := ""
		if v, ok := args["frame_start"].(float64); ok {
			startArg = fmt.Sprintf("\nscene.frame_start = %d", int(v))
		}
		endArg := ""
		if v, ok := args["frame_end"].(float64); ok {
			endArg = fmt.Sprintf("\nscene.frame_end = %d", int(v))
		}
		currentArg := ""
		if v, ok := args["frame_current"].(float64); ok {
			currentArg = fmt.Sprintf("\nscene.frame_set(%d)", int(v))
		}
		fpsArg := ""
		if v, ok := args["fps"].(float64); ok && v > 0 {
			fpsArg = fmt.Sprintf("\nscene.render.fps = %d", int(v))
		}

		if startArg == "" && endArg == "" && currentArg == "" && fpsArg == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "at least one parameter is required"}
		}

		code := fmt.Sprintf(`import bpy
scene = bpy.context.scene
%s%s%s%s
result = {
    "frame_start": scene.frame_start,
    "frame_end": scene.frame_end,
    "frame_current": scene.frame_current,
    "fps": scene.render.fps,
}
`, startArg, endArg, currentArg, fpsArg)
		return execAndFormat(bc, code)
	}
}

func handleSetViewportShading(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		mode := getStringArg(args, "mode")
		if mode == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "mode is required"}
		}

		shadingMap := map[string]string{
			"WIREFRAME": "WIREFRAME",
			"SOLID":      "SOLID",
			"MATERIAL":   "MATERIAL",
			"RENDERED":   "RENDERED",
		}
		shading, ok := shadingMap[mode]
		if !ok {
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown mode: " + mode}
		}

		code := fmt.Sprintf(`import bpy
area = None
for a in bpy.context.screen.areas:
    if a.type == 'VIEW_3D':
        area = a
        break
if area:
    area.spaces.active.shading.type = "%s"
    result = {"shading": "%s"}
else:
    result = {"error": "no 3D viewport found"}
`, shading, shading)
		return execAndFormat(bc, code)
	}
}

func handleExportGLTF(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")
		if filepath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "filepath is required"}
		}
		if !strings.HasSuffix(strings.ToLower(filepath), ".glb") && !strings.HasSuffix(strings.ToLower(filepath), ".gltf") {
			filepath += ".glb"
		}
		code := fmt.Sprintf(`import bpy
bpy.ops.export_scene.gltf(filepath="%s")
result = {"filepath": "%s", "format": "glTF"}
`, filepath, filepath)
		return execAndFormat(bc, code)
	}
}

func handleExportOBJ(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")
		if filepath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "filepath is required"}
		}
		if !strings.HasSuffix(strings.ToLower(filepath), ".obj") {
			filepath += ".obj"
		}
		code := fmt.Sprintf(`import bpy
bpy.ops.wm.obj_export(filepath="%s")
result = {"filepath": "%s", "format": "OBJ"}
`, filepath, filepath)
		return execAndFormat(bc, code)
	}
}

func handleExportFBX(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		filepath := getStringArg(args, "filepath")
		if filepath == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "filepath is required"}
		}
		if !strings.HasSuffix(strings.ToLower(filepath), ".fbx") {
			filepath += ".fbx"
		}
		code := fmt.Sprintf(`import bpy
bpy.ops.export_scene.fbx(filepath="%s")
result = {"filepath": "%s", "format": "FBX"}
`, filepath, filepath)
		return execAndFormat(bc, code)
	}
}

func handleSetupKeyframes(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		raw, ok := args["keyframes"].([]any)
		if !ok || len(raw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "keyframes is required and must be non-empty"}
		}

		// Build Python list of keyframe dicts
		var kfEntries []string
		for _, item := range raw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			objName, _ := m["object"].(string)
			frame, _ := m["frame"].(float64)
			if objName == "" {
				continue
			}
			kfEntries = append(kfEntries, fmt.Sprintf(`{"object": "%s", "frame": %d}`, objName, int(frame)))
		}
		if len(kfEntries) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "no valid keyframes provided"}
		}

		code := fmt.Sprintf(`import bpy
keyframes = [%s]
inserted = 0
objects_set = set()
skipped = set()
for kf in keyframes:
    obj = bpy.data.objects.get(kf["object"])
    if obj is None:
        skipped.add(kf["object"])
        continue
    frame = kf["frame"]
    if "location" in kf:
        obj.location = tuple(kf["location"])
        obj.keyframe_insert(data_path="location", frame=frame)
        inserted += 1
    if "rotation" in kf:
        obj.rotation_euler = tuple(kf["rotation"])
        obj.keyframe_insert(data_path="rotation_euler", frame=frame)
        inserted += 1
    if "scale" in kf:
        obj.scale = tuple(kf["scale"])
        obj.keyframe_insert(data_path="scale", frame=frame)
        inserted += 1
    objects_set.add(kf["object"])
result = {"inserted": inserted, "objects": sorted(objects_set), "skipped": sorted(skipped)}
`, strings.Join(kfEntries, ", "))
		return execAndFormat(bc, code)
	}
}

func handleSetupRigidBody(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		namesRaw, ok := args["object_names"].([]any)
		if !ok || len(namesRaw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_names is required and must be non-empty"}
		}
		names := make([]string, len(namesRaw))
		for i, v := range namesRaw {
			names[i] = fmt.Sprintf("%v", v)
		}

		rbType := "ACTIVE"
		if v, ok := args["rb_type"].(string); ok && v != "" {
			rbType = v
		}
		mass := 1.0
		if v, ok := args["mass"].(float64); ok {
			mass = v
		}
		friction := 0.5
		if v, ok := args["friction"].(float64); ok {
			friction = v
		}
		restitution := 0.3
		if v, ok := args["restitution"].(float64); ok {
			restitution = v
		}
		shape := "CONVEX_HULL"
		if v, ok := args["collision_shape"].(string); ok && v != "" {
			shape = v
		}

		namesPy := strings.Join(func() []string {
			parts := make([]string, len(names))
			for i, n := range names {
				parts[i] = fmt.Sprintf(`"%s"`, n)
			}
			return parts
		}(), ", ")

		code := fmt.Sprintf(`import bpy
obj_names = [%s]
configured = []
skipped = []
for name in obj_names:
    obj = bpy.data.objects.get(name)
    if obj is None:
        skipped.append(name)
        continue
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.rigidbody.object_add()
    obj.rigid_body.type = "%s"
    obj.rigid_body.mass = %g
    obj.rigid_body.friction = %g
    obj.rigid_body.restitution = %g
    obj.rigid_body.collision_shape = "%s"
    obj.select_set(False)
    configured.append(name)
result = {"configured": configured, "skipped": skipped, "rb_type": "%s"}
`, namesPy, rbType, mass, friction, restitution, shape, rbType)
		return execAndFormat(bc, code)
	}
}

func handleSetupFluidDomain(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		domainName := getStringArg(args, "domain_name")
		if domainName == "" {
			domainName = "FluidDomain"
		}

		loc := getFloatSliceArg(args, "location")
		locArg := "0, 0, 2"
		if len(loc) == 3 {
			locArg = fmtFloats(loc)
		}

		size := 4.0
		if v, ok := args["size"].(float64); ok && v > 0 {
			size = v
		}
		resolution := 64
		if v, ok := args["resolution"].(float64); ok && v > 0 {
			resolution = int(v)
		}
		cacheDir := "//fluid_cache"
		if v, ok := args["cache_dir"].(string); ok && v != "" {
			cacheDir = v
		}
		domainType := "LIQUID"
		if v, ok := args["domain_type"].(string); ok && v != "" {
			domainType = v
		}

		code := fmt.Sprintf(`import bpy, bmesh
h = %g / 2.0
verts = [(-h,-h,-h),(h,-h,-h),(h,h,-h),(-h,h,-h),(-h,-h,h),(h,-h,h),(h,h,h),(-h,h,h)]
faces = [(0,1,2,3),(4,5,6,7),(0,1,5,4),(1,2,6,5),(2,3,7,6),(3,0,4,7)]
mesh = bpy.data.meshes.new("%sMesh")
mesh.from_pydata(verts, [], faces)
mesh.update()
obj = bpy.data.objects.new("%s", mesh)
bpy.context.scene.collection.objects.link(obj)
obj.location = (%s)
modifier = obj.modifiers.new(name="Fluid", type="FLUID")
modifier.fluid_type = "DOMAIN"
settings = modifier.domain_settings
settings.domain_type = "%s"
settings.resolution_max = %d
settings.cache_directory = "%s"
modifier.show_viewport = False
obj.display_type = "WIRE"
result = {"domain": obj.name, "resolution": settings.resolution_max, "cache_dir": settings.cache_directory, "domain_type": "%s"}
`, size, domainName, domainName, locArg, domainType, resolution, cacheDir, domainType)
		return execAndFormat(bc, code)
	}
}

func handleSetupFluidInflow(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		flowType := "LIQUID"
		if v, ok := args["flow_type"].(string); ok && v != "" {
			flowType = v
		}
		flowBehavior := "INFLOW"
		if v, ok := args["flow_behavior"].(string); ok && v != "" {
			flowBehavior = v
		}
		particleSize := 0.1
		if v, ok := args["use_particle_size"].(float64); ok && v > 0 {
			particleSize = v
		}

		code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.get("Fluid")
    if mod is None:
        mod = obj.modifiers.new(name="Fluid", type="FLUID")
    mod.fluid_type = "FLOW"
    flow = mod.flow_settings
    flow.flow_type = "%s"
    flow.flow_behavior = "%s"
    flow.use_particle_size = True
    flow.particle_size = %g
    result = {"object": obj.name, "flow_type": "%s", "flow_behavior": "%s"}
`, objName, flowType, flowBehavior, particleSize, flowType, flowBehavior)
		return execAndFormat(bc, code)
	}
}

func handleSetupEffector(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		namesRaw, ok := args["object_names"].([]any)
		if !ok || len(namesRaw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_names is required and must be non-empty"}
		}
		names := make([]string, len(namesRaw))
		for i, v := range namesRaw {
			names[i] = fmt.Sprintf("%v", v)
		}

		effectorType := "COLLISION"
		if v, ok := args["effector_type"].(string); ok && v != "" {
			effectorType = v
		}
		surfaceDistance := 0.01
		if v, ok := args["surface_distance"].(float64); ok && v >= 0 {
			surfaceDistance = v
		}

		namesPy := strings.Join(func() []string {
			parts := make([]string, len(names))
			for i, n := range names {
				parts[i] = fmt.Sprintf(`"%s"`, n)
			}
			return parts
		}(), ", ")

		code := fmt.Sprintf(`import bpy
obj_names = [%s]
configured = []
skipped = []
for name in obj_names:
    obj = bpy.data.objects.get(name)
    if obj is None:
        skipped.append(name)
        continue
    mod = obj.modifiers.get("Fluid")
    if mod is None:
        mod = obj.modifiers.new(name="Fluid", type="FLUID")
    mod.fluid_type = "EFFECTOR"
    eff = mod.effector_settings
    eff.effector_type = "%s"
    eff.surface_distance = %g
    has_collision = any(m.type == "COLLISION" for m in obj.modifiers)
    if not has_collision:
        obj.modifiers.new(name="Collision", type="COLLISION")
    configured.append(name)
result = {"configured": configured, "skipped": skipped}
`, namesPy, effectorType, surfaceDistance)
		return execAndFormat(bc, code)
	}
}

func handleSetupCamera(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		useExisting := getStringArg(args, "use_existing")

		if useExisting != "" {
			loc := getFloatSliceArg(args, "location")
			rot := getFloatSliceArg(args, "rotation")
			focalLength := 50.0
			if v, ok := args["focal_length"].(float64); ok && v > 0 {
				focalLength = v
			}
			setActive := true
			if v, ok := args["set_active"].(bool); ok {
				setActive = v
			}

			locArg := ""
			if len(loc) == 3 {
				locArg = fmt.Sprintf("\ncam.location = (%s)", fmtFloats(loc))
			}
			rotArg := ""
			if len(rot) == 3 {
				rotArg = fmt.Sprintf("\ncam.rotation_euler = (%s)", fmtFloats(rot))
			}
			setActiveArg := ""
			if setActive {
				setActiveArg = "\nbpy.context.scene.camera = cam"
			}

			code := fmt.Sprintf(`import bpy
cam = bpy.data.objects.get("%s")
if cam is None or cam.type != 'CAMERA':
    result = {"error": "camera not found"}
else:
    cam.data.lens = %g%s%s%s
    result = {"name": cam.name, "location": [round(v, 4) for v in cam.location], "focal_length": cam.data.lens, "is_active": bpy.context.scene.camera == cam}
`, useExisting, focalLength, locArg, rotArg, setActiveArg)
			return execAndFormat(bc, code)
		}

		camName := getStringArg(args, "name")
		if camName == "" {
			camName = "Camera"
		}

		loc := getFloatSliceArg(args, "location")
		locArg := "0, 0, 0"
		if len(loc) == 3 {
			locArg = fmtFloats(loc)
		}

		rot := getFloatSliceArg(args, "rotation")
		rotArg := ""
		if len(rot) == 3 {
			rotArg = fmt.Sprintf("\ncam.rotation_euler = (%s)", fmtFloats(rot))
		}

		focalLength := 50.0
		if v, ok := args["focal_length"].(float64); ok && v > 0 {
			focalLength = v
		}

		setActive := true
		if v, ok := args["set_active"].(bool); ok {
			setActive = v
		}
		setActiveArg := ""
		if setActive {
			setActiveArg = "\nbpy.context.scene.camera = cam"
		}

		code := fmt.Sprintf(`import bpy
cam_data = bpy.data.cameras.new("%s")
cam = bpy.data.objects.new("%s", cam_data)
bpy.context.collection.objects.link(cam)
cam.location = (%s)%s
cam.data.lens = %g%s
result = {"name": cam.name, "location": [round(v, 4) for v in cam.location], "focal_length": cam.data.lens, "is_active": bpy.context.scene.camera == cam}
`, camName, camName, locArg, rotArg, focalLength, setActiveArg)
		return execAndFormat(bc, code)
	}
}

func handleManageCollections(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		colsRaw, ok := args["collections"].([]any)
		if !ok || len(colsRaw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "collections is required and must be non-empty"}
		}

		// Build Python list of collection specs
		var colEntries []string
		for _, item := range colsRaw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			if name == "" {
				continue
			}
			parent, _ := m["parent"].(string)
			colorTag, _ := m["color_tag"].(string)

			objNamesRaw, _ := m["objects"].([]any)
			var objPy []string
			for _, o := range objNamesRaw {
				objPy = append(objPy, fmt.Sprintf(`"%v"`, o))
			}
			objectsStr := "[]"
			if len(objPy) > 0 {
				objectsStr = "[" + strings.Join(objPy, ", ") + "]"
			}

			parentStr := "None"
			if parent != "" {
				parentStr = fmt.Sprintf(`"%s"`, parent)
			}
			colorStr := "None"
			if colorTag != "" {
				colorStr = fmt.Sprintf(`"%s"`, colorTag)
			}

			colEntries = append(colEntries, fmt.Sprintf(`{"name": "%s", "objects": %s, "parent": %s, "color_tag": %s}`, name, objectsStr, parentStr, colorStr))
		}
		if len(colEntries) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "no valid collections provided"}
		}

		code := fmt.Sprintf(`import bpy
collections_spec = [%s]
created = []
moved = {}
skipped = []
for spec in collections_spec:
    col_name = spec["name"]
    col = bpy.data.collections.get(col_name)
    if col is None:
        col = bpy.data.collections.new(col_name)
        parent_name = spec.get("parent")
        if parent_name:
            parent = bpy.data.collections.get(parent_name)
            if parent:
                parent.children.link(col)
            else:
                bpy.context.scene.collection.children.link(col)
        else:
            bpy.context.scene.collection.children.link(col)
        created.append(col_name)
    color_tag = spec.get("color_tag")
    if color_tag:
        col.color_tag = color_tag
    moved[col_name] = []
    for obj_name in spec.get("objects", []):
        obj = bpy.data.objects.get(obj_name)
        if obj is None:
            skipped.append(obj_name)
            continue
        for existing_col in obj.users_collection:
            existing_col.objects.unlink(obj)
        col.objects.link(obj)
        moved[col_name].append(obj_name)
result = {"created": created, "moved": moved, "skipped": skipped}
`, strings.Join(colEntries, ", "))
		return execAndFormat(bc, code)
	}
}

// --- helpers ---

func execAndFormat(bc *blender.Client, code string) (any, *Error) {
	result, err := bc.ExecuteCode(code, true)
	if err != nil {
		return nil, &Error{Code: CodeInternalError, Message: err.Error()}
	}
	text := formatExecResult(result)
	return ToolCallResult{
		Content: []ContentBlock{NewTextContent(text)},
		IsError: result.Status == "error",
	}, nil
}

func formatExecResult(r *blender.ExecResult) string {
	var b strings.Builder
	if r.Status == "error" {
		b.WriteString("Error: ")
		b.WriteString(r.Message)
		b.WriteString("\n")
	}
	if r.Result != nil {
		data, _ := json.MarshalIndent(r.Result, "", "  ")
		b.Write(data)
		b.WriteString("\n")
	}
	if r.Stdout != "" {
		b.WriteString("--- stdout ---\n")
		b.WriteString(r.Stdout)
		b.WriteString("\n")
	}
	if r.Stderr != "" {
		b.WriteString("--- stderr ---\n")
		b.WriteString(r.Stderr)
		b.WriteString("\n")
	}
	return b.String()
}
