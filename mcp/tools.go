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
			Description: "List, add, or remove modifiers on an object. action: 'list' (default), 'add', or 'remove'. Accepts any valid Blender modifier type (SUBSURF, MIRROR, ARRAY, SOLIDIFY, BEVEL, BOOLEAN, DECIMATE, REMESH, SMOOTH, LATTICE, ARMATURE, CURVE, SHRINKWRAP, DISPLACE, CLOTH, FLUID, SOFT_BODY, PARTICLE_SYSTEM, etc.).",
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
						"description": "Any valid Blender modifier type (e.g. SUBSURF, MIRROR, ARRAY, SOLIDIFY, BEVEL, BOOLEAN, DECIMATE, REMESH, SMOOTH, LATTICE, ARMATURE, CURVE, SHRINKWRAP, DISPLACE, CLOTH, FLUID, SOFT_BODY, PARTICLE_SYSTEM, GREASEPENCIL_SMOOTH, GREASEPENCIL_MIRROR, etc.)",
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
			Description: "List, add, or remove constraints on an object. action: 'list' (default), 'add', or 'remove'. Accepts any valid Blender constraint type (TRACK_TO, COPY_LOCATION, COPY_ROTATION, COPY_SCALE, LIMIT_DISTANCE, LIMIT_LOCATION, LIMIT_ROTATION, LIMIT_SCALE, CHILD_OF, IK, DAMPED_TRACK, STRETCH_TO, CLAMP_TO, FOLLOW_PATH, TRANSFORM, ACTION, ARMATURE, PIVOT, MAINTAIN_VOLUME, FLOOR, etc.).",
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
						"description": "Any valid Blender constraint type (e.g. TRACK_TO, COPY_LOCATION, COPY_ROTATION, COPY_SCALE, LIMIT_DISTANCE, LIMIT_LOCATION, LIMIT_ROTATION, LIMIT_SCALE, CHILD_OF, IK, DAMPED_TRACK, STRETCH_TO, CLAMP_TO, FOLLOW_PATH, TRANSFORM, ACTION, ARMATURE, PIVOT, MAINTAIN_VOLUME, FLOOR)",
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
			Description: "List, add, or remove physics simulations on an object. action: 'list' (default), 'add', or 'remove'. Supports: RIGID_BODY, CLOTH, FLUID, FORCE_FIELD (with sub-types: FORCE, WIND, VORTEX, MAGNET, RHARBOR, CHARGE, LENNARDJENKINS, TEXTURE, HARMONIC, TURBULENCE, DRAG, SMOKE_FLOW), SOFT_BODY, PARTICLE_SYSTEM, DYNAMIC_PAINT, SIMPLIFY, MESH_SEQUENCE_CACHE.",
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
						"description": "Physics type to add/remove (e.g. RIGID_BODY, CLOTH, FLUID, FORCE_FIELD, SOFT_BODY, PARTICLE_SYSTEM, DYNAMIC_PAINT, SIMPLIFY, MESH_SEQUENCE_CACHE)",
					},
					"force_field_type": map[string]any{
						"type":        "string",
						"description": "Force field sub-type when physics_type is FORCE_FIELD",
						"enum":        []string{"FORCE", "WIND", "VORTEX", "MAGNET", "RHARBOR", "CHARGE", "LENNARDJENKINS", "TEXTURE", "HARMONIC", "TURBULENCE", "DRAG", "SMOKE_FLOW"},
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
		{
			Name:        "set_cursor",
			Description: "Set the 3D cursor location and optionally rotation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "array",
						"description": "3D cursor [x, y, z] position",
						"items":       map[string]any{"type": "number"},
					},
					"rotation": map[string]any{
						"type":        "array",
						"description": "3D cursor [rx, ry, rz] rotation in degrees",
						"items":       map[string]any{"type": "number"},
					},
				},
			},
		},
		{
			Name:        "set_snap",
			Description: "Configure snap settings: snap type, target, and toggle.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"use_snap": map[string]any{
						"type":        "boolean",
						"description": "Enable/disable snapping",
					},
					"snap_target": map[string]any{
						"type":        "string",
						"description": "Snap target",
						"enum":        []string{"CLOSEST", "CENTER", "MEDIAN", "ACTIVE"},
					},
					"snap_element": map[string]any{
						"type":        "string",
						"description": "Snap element type",
						"enum":        []string{"VERTEX", "EDGE", "FACE", "VOLUME", "INCREMENT", "GRID"},
					},
				},
			},
		},
		{
			Name:        "edit_mesh_data",
			Description: "Edit mesh data directly: add/remove vertices, edges, faces. Returns mesh stats.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object to edit",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "Action to perform",
						"enum":        []string{"stats", "add_vertices", "remove_vertices", "Triangulate", "RemoveDoubles", "RecalculateNormals"},
					},
					"vertices": map[string]any{
						"type":        "array",
						"description": "Vertices to add [[x,y,z], ...]",
						"items":       map[string]any{"type": "array", "items": map[string]any{"type": "number"}},
					},
					"indices": map[string]any{
						"type":        "array",
						"description": "Vertex indices to remove",
						"items":       map[string]any{"type": "integer"},
					},
				},
				"required": []string{"object_name", "action"},
			},
		},
		{
			Name:        "set_particle_system",
			Description: "Add or configure a particle system on an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object to add particles to",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "Action to perform",
						"enum":        []string{"add", "remove", "configure", "list"},
					},
					"particle_type": map[string]any{
						"type":        "string",
						"description": "Particle type: EMITTER or HAIR",
						"enum":        []string{"EMITTER", "HAIR"},
					},
					"count": map[string]any{
						"type":        "integer",
						"description": "Number of particles",
					},
					"seed": map[string]any{
						"type":        "integer",
						"description": "Random seed",
					},
					"lifetime": map[string]any{
						"type":        "number",
						"description": "Particle lifetime in frames",
					},
				},
				"required": []string{"object_name", "action"},
			},
		},
		{
			Name:        "set_render_passes",
			Description: "Enable or disable render passes for the active view layer.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"passes": map[string]any{
						"type":        "object",
						"description": "Pass name -> enabled mapping, e.g. {\"use_pass_z\": true, \"use_pass_normal\": true}",
						"additionalProperties": map[string]any{"type": "boolean"},
					},
				},
				"required": []string{"passes"},
			},
		},
		{
			Name:        "create_object",
			Description: "Create any Blender object type: EMPTY, ARMATURE, CURVE, CURVES, FONT, GREASEPENCIL, LATTICE, LIGHT_PROBE, META, POINTCLOUD, SPEAKER, SURFACE, VOLUME. For MESH, use create_mesh_object instead.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_type": map[string]any{
						"type":        "string",
						"description": "Object type to create",
						"enum":        []string{"EMPTY", "ARMATURE", "CURVE", "CURVES", "FONT", "GREASEPENCIL", "LATTICE", "LIGHT_PROBE", "META", "POINTCLOUD", "SPEAKER", "SURFACE", "VOLUME"},
					},
					"name": map[string]any{
						"type":        "string",
						"description": "Object name",
					},
					"location": map[string]any{
						"type":        "array",
						"description": "Location [x, y, z]",
						"items":       map[string]any{"type": "number"},
					},
					"empty_display_type": map[string]any{
						"type":        "string",
						"description": "Empty display type (EMPTY only)",
						"enum":        []string{"PLAIN_AXES", "ARROWS", "SINGLE_ARROW", "CIRCLE", "CUBE", "SPHERE", "CONE", "MATMESH"},
					},
					"curve_type": map[string]any{
						"type":        "string",
						"description": "Curve type: BEZIER or NURBS (CURVE only)",
						"enum":        []string{"BEZIER", "NURBS"},
					},
					"font_text": map[string]any{
						"type":        "string",
						"description": "Text content (FONT only)",
					},
				},
				"required": []string{"object_type"},
			},
		},
		{
			Name:        "grease_pencil_manage",
			Description: "Manage Grease Pencil objects: draw strokes, list/apply modifiers, manage layers, fill, extrude, delete, and other GP operations.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Grease Pencil object name",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "Action to perform",
						"enum": []string{
							"list_layers", "add_layer", "remove_layer",
							"list_modifiers", "add_modifier", "remove_modifier",
							"draw_stroke", "fill", "extrude", "delete",
							"dissolve", "duplicate", "clean_loose",
							"convert_curve_type", "set_cyclic",
							"list_frames", "add_frame", "remove_frame",
							"active_frame_delete", "delete_frame",
							"brush_stroke", "caps_set", "copy",
							"delete_breakdown", "erase_box", "erase_lasso",
							"frame_clean_duplicate", "bake_grease_pencil_animation",
						},
					},
					"layer_name": map[string]any{
						"type":        "string",
						"description": "Layer name (for layer actions)",
					},
					"modifier_name": map[string]any{
						"type":        "string",
						"description": "Modifier name (for modifier actions)",
					},
					"modifier_type": map[string]any{
						"type":        "string",
						"description": "GP modifier type (for add_modifier)",
						"enum": []string{
							"GREASE_PENCIL_ARRAY", "GREASE_PENCIL_BUILD", "GREASE_PENCIL_MIRROR",
							"GREASE_PENCIL_MULTIPLY", "GREASE_PENCIL_NOISE", "GREASE_PENCIL_OFFSET",
							"GREASE_PENCIL_SMOOTH", "GREASE_PENCIL_SUBDIV", "GREASE_PENCIL_ENVELOPE",
							"GREASE_PENCIL_OUTLINE", "GREASE_PENCIL_HOOK", "GREASE_PENCIL_LATTICE",
							"GREASE_PENCIL_DASH", "GREASE_PENCIL_ARMATURE", "GREASE_PENCIL_SHRINKWRAP",
							"GREASE_PENCIL_SIMPLIFY", "GREASE_PENCIL_THICKNESS", "GREASE_PENCIL_LENGTH",
							"GREASE_PENCIL_COLOR", "GREASE_PENCIL_TINT", "GREASE_PENCIL_OPACITY",
							"GREASE_PENCIL_TEXTURE", "GREASE_PENCIL_TIME",
							"GREASE_PENCIL_VERTEX_WEIGHT_PROXIMITY", "GREASE_PENCIL_VERTEX_WEIGHT_ANGLE",
						},
					},
					"points": map[string]any{
						"type":        "array",
						"description": "Stroke points [[x,y,z], ...] (for draw_stroke)",
						"items":       map[string]any{"type": "array", "items": map[string]any{"type": "number"}},
					},
					"pressure": map[string]any{
						"type":        "number",
						"description": "Pen pressure (for draw_stroke)",
					},
					"curve_type": map[string]any{
						"type":        "string",
						"description": "Target curve type (for convert_curve_type)",
						"enum":        []string{"BEZIER", "NURBS", "POLY"},
					},
				},
				"required": []string{"object_name", "action"},
			},
		},
		{
			Name:        "compositor_nodes",
			Description: "Manage compositor node tree: add/remove/connect nodes, list available node types.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Action to perform",
						"enum":        []string{"list_types", "list_nodes", "add_node", "remove_node", "connect_nodes"},
					},
					"node_type": map[string]any{
						"type":        "string",
						"description": "Compositor node class name (e.g. CompositorNodeRGB, CompositorNodeBlur)",
					},
					"node_name": map[string]any{
						"type":        "string",
						"description": "Instance name of a specific node",
					},
					"from_node": map[string]any{
						"type":        "string",
						"description": "Source node name (for connect_nodes)",
					},
					"from_socket": map[string]any{
						"type":        "string",
						"description": "Source socket name (for connect_nodes)",
					},
					"to_node": map[string]any{
						"type":        "string",
						"description": "Destination node name (for connect_nodes)",
					},
					"to_socket": map[string]any{
						"type":        "string",
						"description": "Destination socket name (for connect_nodes)",
					},
				},
				"required": []string{"action"},
			},
		},
		{
			Name:        "node_wrangler_ops",
			Description: "Node Wrangler add-on shortcuts: preview links, swap nodes, mix nodes, collapse/expand, frames, textures setup. Requires Shader Editor to be the active context.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Node Wrangler action",
						"enum": []string{
							"preview_link", "swap_nodes", "mix_nodes",
							"collapse_all", "expand_all", "frame_selected",
							"add_texture_setup", "connect_viewer",
							"disconnect_viewer", "node_switch",
						},
					},
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object with material (for material node actions)",
					},
					"node_tree_type": map[string]any{
						"type":        "string",
						"description": "Node tree type: MATERIAL, WORLD, or COMpositor",
						"enum":        []string{"MATERIAL", "WORLD", "COMpositor"},
					},
				},
				"required": []string{"action"},
			},
		},
		{
			Name:        "pose_library_ops",
			Description: "Pose Library operations: create asset library, save pose, apply pose, list poses.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Pose library action",
						"enum":        []string{"save_pose", "apply_pose", "list_poses", "create_library"},
					},
					"armature_name": map[string]any{
						"type":        "string",
						"description": "Armature object name",
					},
					"pose_name": map[string]any{
						"type":        "string",
						"description": "Pose name",
					},
					"library_path": map[string]any{
						"type":        "string",
						"description": "Asset library path (for create_library)",
					},
				},
				"required": []string{"action"},
			},
		},
		{
			Name:        "rigify_ops",
			Description: "Rigify add-on operations: generate rig, generate metarig, list rig types.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"description": "Rigify action",
						"enum":        []string{"generate_rig", "add_metarig", "list_metarigs", "list_rig_types"},
					},
					"armature_name": map[string]any{
						"type":        "string",
						"description": "Armature/metarig name",
					},
					"metarig_type": map[string]any{
						"type":        "string",
						"description": "Metarig template type (for add_metarig)",
						"enum": []string{
							"basic", "human", "quadruped", "bird", "cat",
							"horse", "monkey", "shark", "wolf",
							"arm.finger", "leg.plane.02", "spine.basic.01",
							"spine.reptile.01", "head.basic.01",
						},
					},
				},
				"required": []string{"action"},
			},
		},
		{
			Name:        "manage_shader_nodes",
			Description: "Manage material shader node tree: add/remove/connect nodes, list available node types. Works on the active material of an object.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"object_name": map[string]any{
						"type":        "string",
						"description": "Object with material to edit",
					},
					"action": map[string]any{
						"type":        "string",
						"description": "Action to perform",
						"enum":        []string{"list_types", "list_nodes", "add_node", "remove_node", "connect_nodes"},
					},
					"node_type": map[string]any{
						"type":        "string",
						"description": "Shader node class name (e.g. ShaderNodeBsdfPrincipled, ShaderNodeTexImage, ShaderNodeMix)",
					},
					"node_name": map[string]any{
						"type":        "string",
						"description": "Instance name of a specific node",
					},
					"from_node": map[string]any{
						"type":        "string",
						"description": "Source node name (for connect_nodes)",
					},
					"from_socket": map[string]any{
						"type":        "string",
						"description": "Source socket name (for connect_nodes)",
					},
					"to_node": map[string]any{
						"type":        "string",
						"description": "Destination node name (for connect_nodes)",
					},
					"to_socket": map[string]any{
						"type":        "string",
						"description": "Destination socket name (for connect_nodes)",
					},
				},
				"required": []string{"object_name", "action"},
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
		{ToolDefs()[48], handleSetRenderEngine(bc)},
		{ToolDefs()[49], handleSetRenderFormat(bc)},
		{ToolDefs()[50], handleSetWorldEnvironment(bc)},
		{ToolDefs()[51], handleSetCursor(bc)},
		{ToolDefs()[52], handleSetSnap(bc)},
		{ToolDefs()[53], handleEditMeshData(bc)},
		{ToolDefs()[54], handleSetParticleSystem(bc)},
		{ToolDefs()[55], handleSetRenderPasses(bc)},
		{ToolDefs()[56], handleCreateObject(bc)},
		{ToolDefs()[57], handleGreasePencilManage(bc)},
		{ToolDefs()[58], handleCompositorNodes(bc)},
		{ToolDefs()[59], handleNodeWranglerOps(bc)},
		{ToolDefs()[60], handlePoseLibraryOps(bc)},
		{ToolDefs()[61], handleRigifyOps(bc)},
		{ToolDefs()[62], handleManageShaderNodes(bc)},
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

// helper to convert any numeric value to float64
func toFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

// escapePyStr escapes a string for safe embedding in Python single-quoted strings.
// Handles backslashes (Windows paths) and single quotes.
func escapePyStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
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
			filterClause = fmt.Sprintf(`
    if obj.type != "%s":
        continue`, filterType)
		}

		code := fmt.Sprintf(`import bpy, json
scene = bpy.context.scene
objects = []
for obj in scene.objects:%s
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
			nameArg = fmt.Sprintf("obj.name = '%s'\n", escapePyStr(name))
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
			// No data API path for Suzanne — use bpy.ops but DON'T delete scene
			locParam := ""
			if len(loc) == 3 {
				locParam = fmt.Sprintf("location=(%s)", fmtFloats(loc))
			}
			code = fmt.Sprintf(`import bpy
bpy.ops.mesh.primitive_monkey_add(%s)
obj = bpy.context.active_object
if obj:
    %sresult = {"name": obj.name, "type": obj.type, "location": [round(v, 4) for v in obj.location]}
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
if scene.camera is None:
    cam = bpy.data.cameras.new("AutoCamera")
    cam_obj = bpy.data.objects.new("AutoCamera", cam)
    scene.collection.objects.link(cam_obj)
    scene.camera = cam_obj
    cam_obj.location = (0, -10, 5)
    cam_obj.rotation_euler = (1.1, 0, 0)
scene.render.filepath = '%s'%s%s
bpy.ops.render.render(write_still=True)
result = {"rendered": True, "output_path": '%s'}
`, escapePyStr(outputPath), resXArg, resYArg, escapePyStr(outputPath))
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
			importFn = fmt.Sprintf(`bpy.ops.import_scene.gltf(filepath='%s')`, escapePyStr(filepath))
		case ".fbx":
			importFn = fmt.Sprintf(`bpy.ops.import_scene.fbx(filepath='%s')`, escapePyStr(filepath))
		case ".obj":
			importFn = fmt.Sprintf(`bpy.ops.wm.obj_import(filepath='%s')`, escapePyStr(filepath))
		case ".stl":
			importFn = fmt.Sprintf(`bpy.ops.wm.stl_import(filepath='%s')`, escapePyStr(filepath))
		case ".ply":
			importFn = fmt.Sprintf(`bpy.ops.wm.ply_import(filepath='%s')`, escapePyStr(filepath))
		case ".bvh":
			importFn = fmt.Sprintf(`bpy.ops.import_anim.bvh(filepath='%s')`, escapePyStr(filepath))
		case ".svg":
			importFn = fmt.Sprintf(`bpy.ops.import_curve.svg(filepath='%s')`, escapePyStr(filepath))
		case ".vrm":
			importFn = fmt.Sprintf(`bpy.ops.import_scene.vrm(filepath='%s')`, escapePyStr(filepath))
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
			exportFn = fmt.Sprintf(`bpy.ops.export_scene.gltf(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		case ".fbx":
			exportFn = fmt.Sprintf(`bpy.ops.export_scene.fbx(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		case ".obj":
			exportFn = fmt.Sprintf(`bpy.ops.wm.obj_export(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		case ".stl":
			exportFn = fmt.Sprintf(`bpy.ops.wm.stl_export(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		case ".ply":
			exportFn = fmt.Sprintf(`bpy.ops.wm.ply_export(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		case ".bvh":
			exportFn = fmt.Sprintf(`bpy.ops.export_anim.bvh(filepath='%s'%s)`, escapePyStr(filepath), selArg)
		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unsupported file type: " + ext}
		}

		code := fmt.Sprintf(`import bpy
%s
result = {"exported": '%s'}
`, exportFn, escapePyStr(filepath))
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
			locArg = fmt.Sprintf("location=(%s)", fmtFloats(loc))
		}
		energyArg := ""
		if v, ok := args["energy"].(float64); ok && v > 0 {
			energyArg = fmt.Sprintf("\n    light.data.energy = %g", v)
		}
		color := getFloatSliceArg(args, "color")
		colorArg := ""
		if len(color) >= 3 {
			colorArg = fmt.Sprintf("\n    light.data.color = (%s)", fmtFloats(color[:3]))
		}
		name := getStringArg(args, "name")
		nameArg := ""
		if name != "" {
			nameArg = fmt.Sprintf("\n    light.name = '%s'", escapePyStr(name))
		}

		locSep := ""
		if locArg != "" {
			locSep = ", "
		}
		code := fmt.Sprintf(`import bpy
bpy.ops.object.light_add(type="%s"%s%s)
light = bpy.context.active_object
if light:%s%s%s
result = {"name": light.name, "type": light.data.type} if light else {"error": "failed to add light"}
`, lightType, locSep, locArg, nameArg, energyArg, colorArg)
		return execAndFormat(bc, code)
	}
}

func handleAddCamera(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		loc := getFloatSliceArg(args, "location")
		locArg := ""
		if len(loc) == 3 {
			locArg = fmt.Sprintf("location=(%s)", fmtFloats(loc))
		}
		rot := getFloatSliceArg(args, "rotation")
		rotArg := ""
		if len(rot) == 3 {
			rotArg = fmt.Sprintf("\n    cam.rotation_euler = (%s)", fmtFloats(rot))
		}
		name := getStringArg(args, "name")
		nameArg := ""
		if name != "" {
			nameArg = fmt.Sprintf("\n    cam.name = '%s'", escapePyStr(name))
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
try:
    bpy.ops.ed.undo()
    result = {"undone": True}
except RuntimeError as e:
    result = {"error": str(e), "hint": "undo requires GUI context; try execute_code directly"}`
		return execAndFormat(bc, code)
	}
}

func handleRedo(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		code := `import bpy
try:
    bpy.ops.ed.redo()
    result = {"redone": True}
except RuntimeError as e:
    result = {"error": str(e), "hint": "redo requires GUI context; try execute_code directly"}`
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
        "children": [c.name for c in col.children],
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
os.makedirs('%s', exist_ok=True)
scene.render.filepath = os.path.join('%s', 'frame_####')
%s%s
bpy.ops.render.render(animation=True)
result = {"rendered": True, "output_dir": '%s'}
`, escapePyStr(outputDir), escapePyStr(outputDir), startArg, endArg, escapePyStr(outputDir))
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
mat = bpy.data.materials.get('%s')
if mat is None:
    result = {"error": "material not found"}
else:
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes.get("Principled BSDF")
    if bsdf is None:
        result = {"error": "no Principled BSDF node found"}
    else:
        img_node = mat.node_tree.nodes.new("ShaderNodeTexImage")
        img_node.image = bpy.data.images.load('%s')
        mat.node_tree.links.new(img_node.outputs["Color"], bsdf.inputs["Base Color"])
        result = {"material": mat.name, "texture": '%s'}
`, escapePyStr(matName), escapePyStr(imagePath), escapePyStr(imagePath))
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
            def safe_val(socket):
                if not hasattr(socket, 'default_value') or socket.default_value is None:
                    return None
                try:
                    return list(socket.default_value)
                except:
                    return str(socket.default_value)
            inputs = [{"name": inp.name, "type": inp.type, "default_value": safe_val(inp)} for inp in n.inputs]
            outputs = [{"name": out.name, "type": out.type, "default_value": safe_val(out)} for out in n.outputs]
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
		objCode := ""
		if objName != "" {
			objCode = fmt.Sprintf(`obj = bpy.data.objects.get('%s')
if obj is None:
    result = {"error": "object not found"}
`, escapePyStr(objName))
		} else {
			objCode = `obj = bpy.context.active_object
if obj is None:
    result = {"error": "no active object"}
`
		}

		code := fmt.Sprintf(`import bpy
%s
keyframes = {}
if obj.animation_data and obj.animation_data.action:
    action = obj.animation_data.action
    fc_list = []
    if action.is_action_legacy:
        fc_list = list(action.fcurves)
    else:
        for layer in action.layers:
            for strip in layer.strips:
                for cb in strip.channelbags:
                    fc_list.extend(list(cb.fcurves))
    for fc in fc_list:
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
`, objCode)
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
    if obj.modifiers.get("Dynamic Paint"):
        phys.append({"type": "DYNAMIC_PAINT"})
    if obj.modifiers.get("Simplify"):
        phys.append({"type": "SIMPLIFY"})
    if obj.modifiers.get("Mesh Sequence Cache"):
        phys.append({"type": "MESH_SEQUENCE_CACHE"})
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
				ffType := getStringArg(args, "force_field_type")
				if ffType == "" {
					ffType = "WIND"
				}
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.effector_add(type="%s")
    result = {"added": "FORCE_FIELD", "field_type": "%s", "object": obj.name}
`, objName, ffType, ffType)
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
			case "DYNAMIC_PAINT":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Dynamic Paint", type="DYNAMIC_PAINT")
    result = {"added": "DYNAMIC_PAINT", "object": obj.name}
`, objName)
			case "SIMPLIFY":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Simplify", type="SIMPLIFY")
    result = {"added": "SIMPLIFY", "object": obj.name}
`, objName)
			case "MESH_SEQUENCE_CACHE":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mod = obj.modifiers.new(name="Mesh Sequence Cache", type="MESH_SEQUENCE_CACHE")
    result = {"added": "MESH_SEQUENCE_CACHE", "object": obj.name}
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
			case "FORCE_FIELD":
				code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    if obj.field:
        obj.field = None
        result = {"removed": "FORCE_FIELD", "object": obj.name}
    else:
        result = {"error": "no force field on object"}
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
`, objName, physType, strings.ToUpper(physType[:1])+strings.ToLower(physType[1:]), physType)
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
abs_path = os.path.abspath('%s')
if not abs_path.endswith(".blend"):
    abs_path += ".blend"
bpy.ops.wm.save_as_mainfile(filepath=abs_path, compress=True, relative_remap=True)
result = {"filepath": abs_path, "compressed": True}
`, escapePyStr(filepath))
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
bpy.ops.export_scene.gltf(filepath='%s')
result = {"filepath": '%s', "format": "glTF"}
`, escapePyStr(filepath), escapePyStr(filepath))
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
bpy.ops.wm.obj_export(filepath='%s')
result = {"filepath": '%s', "format": "OBJ"}
`, escapePyStr(filepath), escapePyStr(filepath))
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
bpy.ops.export_scene.fbx(filepath='%s')
result = {"filepath": '%s', "format": "FBX"}
`, escapePyStr(filepath), escapePyStr(filepath))
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
			// Build full JSON dict preserving location/rotation/scale arrays
			parts := []string{
				fmt.Sprintf(`"object": "%s"`, objName),
				fmt.Sprintf(`"frame": %d`, int(frame)),
			}
			if loc, ok := m["location"].([]any); ok && len(loc) == 3 {
				parts = append(parts, fmt.Sprintf(`"location": [%g,%g,%g]`, toFloat(loc[0]), toFloat(loc[1]), toFloat(loc[2])))
			}
			if rot, ok := m["rotation"].([]any); ok && len(rot) == 3 {
				parts = append(parts, fmt.Sprintf(`"rotation": [%g,%g,%g]`, toFloat(rot[0]), toFloat(rot[1]), toFloat(rot[2])))
			}
			if sc, ok := m["scale"].([]any); ok && len(sc) == 3 {
				parts = append(parts, fmt.Sprintf(`"scale": [%g,%g,%g]`, toFloat(sc[0]), toFloat(sc[1]), toFloat(sc[2])))
			}
			kfEntries = append(kfEntries, fmt.Sprintf(`{%s}`, strings.Join(parts, ", ")))
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

func handleSetRenderEngine(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		engine := getStringArg(args, "engine")
		if engine == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "engine is required"}
		}
		code := fmt.Sprintf(`import bpy
bpy.context.scene.render.engine = "%s"
result = {"engine": bpy.context.scene.render.engine}
`, engine)
		return execAndFormat(bc, code)
	}
}

func handleSetRenderFormat(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		parts := []string{}
		if v := getStringArg(args, "file_format"); v != "" {
			parts = append(parts, fmt.Sprintf(`r.image_settings.file_format = "%s"`, v))
		}
		if v := getStringArg(args, "color_mode"); v != "" {
			parts = append(parts, fmt.Sprintf(`r.image_settings.color_mode = "%s"`, v))
		}
		if v := getStringArg(args, "color_depth"); v != "" {
			parts = append(parts, fmt.Sprintf(`r.image_settings.color_depth = "%s"`, v))
		}
		if v, ok := args["resolution_x"].(float64); ok {
			parts = append(parts, fmt.Sprintf(`r.resolution_x = %d`, int(v)))
		}
		if v, ok := args["resolution_y"].(float64); ok {
			parts = append(parts, fmt.Sprintf(`r.resolution_y = %d`, int(v)))
		}
		if v, ok := args["resolution_percentage"].(float64); ok {
			parts = append(parts, fmt.Sprintf(`r.resolution_percentage = %d`, int(v)))
		}
		if len(parts) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "at least one parameter is required"}
		}
		code := fmt.Sprintf(`import bpy
r = bpy.context.scene.render
%s
result = {
    "file_format": r.image_settings.file_format,
    "color_mode": r.image_settings.color_mode,
    "resolution_x": r.resolution_x,
    "resolution_y": r.resolution_y,
    "resolution_percentage": r.resolution_percentage,
}
`, strings.Join(parts, "\n"))
		return execAndFormat(bc, code)
	}
}

func handleSetWorldEnvironment(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		parts := []string{}
		if color := getFloatSliceArg(args, "color"); len(color) >= 3 {
			parts = append(parts, fmt.Sprintf(`bg_node = world.node_tree.nodes.get("Background")
if bg_node:
    bg_node.inputs["Color"].default_value = (%.4f, %.4f, %.4f, 1.0)`, color[0], color[1], color[2]))
		}
		if v, ok := args["strength"].(float64); ok {
			parts = append(parts, fmt.Sprintf(`bg_node = world.node_tree.nodes.get("Background")
if bg_node:
    bg_node.inputs["Strength"].default_value = %.4f`, v))
		}
		if texPath := getStringArg(args, "texture_path"); texPath != "" {
			parts = append(parts, fmt.Sprintf(`import os
tex_node = world.node_tree.nodes.new('ShaderNodeTexEnvironment')
tex_node.image = bpy.data.images.load(r"%s")
bg_node = world.node_tree.nodes.get("Background")
if bg_node:
    world.node_tree.links.new(tex_node.outputs["Color"], bg_node.inputs["Color"])`, texPath))
		}
		if len(parts) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "at least one parameter is required"}
		}
		code := fmt.Sprintf(`import bpy
world = bpy.context.scene.world
if world is None:
    world = bpy.data.worlds.new("World")
    bpy.context.scene.world = world
if not world.use_nodes:
    world.use_nodes = True
%s
result = {"world": world.name, "use_nodes": world.use_nodes}
`, strings.Join(parts, "\n"))
		return execAndFormat(bc, code)
	}
}

func handleSetCursor(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		parts := []string{}
		if loc := getFloatSliceArg(args, "location"); len(loc) >= 3 {
			parts = append(parts, fmt.Sprintf(`bpy.context.scene.cursor.location = (%.6f, %.6f, %.6f)`, loc[0], loc[1], loc[2]))
		}
		if rot := getFloatSliceArg(args, "rotation"); len(rot) >= 3 {
			parts = append(parts, fmt.Sprintf(`import math
bpy.context.scene.cursor.rotation_euler = (math.radians(%.4f), math.radians(%.4f), math.radians(%.4f))`, rot[0], rot[1], rot[2]))
		}
		if len(parts) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "at least one of location or rotation is required"}
		}
		code := fmt.Sprintf(`import bpy
%s
c = bpy.context.scene.cursor
result = {"location": list(c.location), "rotation": [round(math.degrees(r), 2) for c_index, r in enumerate(c.rotation_euler)] if hasattr(c, 'rotation_euler') else []}
`, strings.Join(parts, "\n"))
		return execAndFormat(bc, code)
	}
}

func handleSetSnap(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		parts := []string{}
		if v, ok := args["use_snap"].(bool); ok {
			parts = append(parts, fmt.Sprintf(`ts.use_snap = %v`, v))
		}
		if v := getStringArg(args, "snap_target"); v != "" {
			parts = append(parts, fmt.Sprintf(`ts.snap_target = "%s"`, v))
		}
		if v := getStringArg(args, "snap_element"); v != "" {
			parts = append(parts, fmt.Sprintf(`ts.snap_element = "%s"`, v))
		}
		if len(parts) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "at least one parameter is required"}
		}
		code := fmt.Sprintf(`import bpy
ts = bpy.context.scene.tool_settings
%s
result = {
    "use_snap": ts.use_snap,
    "snap_target": ts.snap_target,
    "snap_element": ts.snap_element,
}
`, strings.Join(parts, "\n"))
		return execAndFormat(bc, code)
	}
}

func handleEditMeshData(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		action := getStringArg(args, "action")
		if objName == "" || action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name and action are required"}
		}

		switch action {
		case "stats":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type != 'MESH':
    result = {"error": "object not found or not a mesh"}
else:
    mesh = obj.data
    result = {
        "object": obj.name,
        "vertices": len(mesh.vertices),
        "edges": len(mesh.edges),
        "polygons": len(mesh.polygons),
        "loops": len(mesh.loops),
    }
`, objName)
			return execAndFormat(bc, code)

		case "add_vertices":
			vertsRaw, ok := args["vertices"].([]any)
			if !ok || len(vertsRaw) == 0 {
				return nil, &Error{Code: CodeInvalidParams, Message: "vertices array is required"}
			}
			vertStrs := []string{}
			for _, v := range vertsRaw {
				if arr, ok := v.([]any); ok && len(arr) >= 3 {
					vertStrs = append(vertStrs, fmt.Sprintf(`(%.6f, %.6f, %.6f)`, toFloat(arr[0]), toFloat(arr[1]), toFloat(arr[2])))
				}
			}
			code := fmt.Sprintf(`import bpy, bmesh
obj = bpy.data.objects.get("%s")
if obj is None or obj.type != 'MESH':
    result = {"error": "object not found or not a mesh"}
else:
    bm = bmesh.new()
    bm.from_mesh(obj.data)
    new_verts = [bm.verts.new(v) for v in [%s]]
    bm.to_mesh(obj.data)
    bm.free()
    obj.data.update()
    result = {"added_vertices": len(new_verts), "total_vertices": len(obj.data.vertices)}
`, objName, strings.Join(vertStrs, ", "))
			return execAndFormat(bc, code)

		case "remove_vertices":
			indicesRaw, ok := args["indices"].([]any)
			if !ok || len(indicesRaw) == 0 {
				return nil, &Error{Code: CodeInvalidParams, Message: "indices array is required"}
			}
			indices := []string{}
			for _, idx := range indicesRaw {
				indices = append(indices, fmt.Sprintf(`%d`, int(toFloat(idx))))
			}
			code := fmt.Sprintf(`import bpy, bmesh
obj = bpy.data.objects.get("%s")
if obj is None or obj.type != 'MESH':
    result = {"error": "object not found or not a mesh"}
else:
    bm = bmesh.new()
    bm.from_mesh(obj.data)
    bm.verts.ensure_lookup_table()
    to_remove = [bm.verts[i] for i in [%s] if i < len(bm.verts)]
    bmesh.ops.delete(bm, geom=to_remove, context='VERTS')
    bm.to_mesh(obj.data)
    bm.free()
    obj.data.update()
    result = {"removed_vertices": len(to_remove), "total_vertices": len(obj.data.vertices)}
`, objName, strings.Join(indices, ", "))
			return execAndFormat(bc, code)

		case "Triangulate", "RemoveDoubles", "RecalculateNormals":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type != 'MESH':
    result = {"error": "object not found or not a mesh"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.mode_set(mode='EDIT')
    bpy.ops.mesh.%s()
    bpy.ops.object.mode_set(mode='OBJECT')
    result = {"action": "%s", "object": obj.name}
`, objName, strings.ToLower(action), action)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleSetParticleSystem(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		action := getStringArg(args, "action")
		if objName == "" || action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name and action are required"}
		}

		switch action {
		case "list":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    systems = []
    for ps in obj.particle_systems:
        settings = ps.settings
        systems.append({
            "name": ps.name,
            "type": settings.type,
            "count": settings.count,
            "seed": settings.seed,
            "lifetime": settings.lifetime,
        })
    result = {"object": obj.name, "particle_systems": systems, "count": len(systems)}
`, objName)
			return execAndFormat(bc, code)

		case "add":
			particleType := getStringArg(args, "particle_type")
			if particleType == "" {
				particleType = "EMITTER"
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.particle_system_add()
    ps = obj.particle_systems.active
    ps.settings.type = "%s"
`, objName, particleType)
			if v, ok := args["count"].(float64); ok {
				code += fmt.Sprintf(`    ps.settings.count = %d
`, int(v))
			}
			if v, ok := args["lifetime"].(float64); ok {
				code += fmt.Sprintf(`    ps.settings.lifetime = %.1f
`, v)
			}
			if v, ok := args["seed"].(float64); ok {
				code += fmt.Sprintf(`    ps.settings.seed = %d
`, int(v))
			}
			code += `    result = {"object": obj.name, "particle_system": ps.name, "type": ps.settings.type}
`
			return execAndFormat(bc, code)

		case "remove":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
elif len(obj.particle_systems) == 0:
    result = {"error": "no particle systems"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.object.particle_system_remove()
    result = {"object": obj.name, "remaining": len(obj.particle_systems)}
`, objName)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleSetRenderPasses(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		passesRaw, ok := args["passes"].(map[string]any)
		if !ok || len(passesRaw) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "passes is required and must be non-empty"}
		}
		setters := []string{}
		for key, val := range passesRaw {
			if b, ok := val.(bool); ok {
				setters = append(setters, fmt.Sprintf(`vl.%s = %v`, key, b))
			}
		}
		if len(setters) == 0 {
			return nil, &Error{Code: CodeInvalidParams, Message: "no valid pass settings provided"}
		}
		code := fmt.Sprintf(`import bpy
vl = bpy.context.view_layer
%s
result = {k: getattr(vl, k) for k in dir(vl) if k.startswith('use_pass_')}
`, strings.Join(setters, "\n"))
		return execAndFormat(bc, code)
	}
}

func handleCreateObject(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objType := getStringArg(args, "object_type")
		if objType == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_type is required"}
		}
		objName := getStringArg(args, "name")
		if objName == "" {
			objName = "Object"
		}
		loc := getFloatSliceArg(args, "location")
		locStr := "(0, 0, 0)"
		if len(loc) >= 3 {
			locStr = fmt.Sprintf("(%.6f, %.6f, %.6f)", loc[0], loc[1], loc[2])
		}

		var code string
		switch objType {
		case "EMPTY":
			displayType := getStringArg(args, "empty_display_type")
			if displayType == "" {
				displayType = "PLAIN_AXES"
			}
			code = fmt.Sprintf(`import bpy
obj = bpy.data.objects.new("%s", None)
obj.empty_display_type = "%s"
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": obj.type, "display_type": obj.empty_display_type}
`, objName, displayType, locStr)

		case "ARMATURE":
			code = fmt.Sprintf(`import bpy
arm = bpy.data.armatures.new("%s_data")
obj = bpy.data.objects.new("%s", arm)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
bpy.ops.object.mode_set(mode='EDIT')
bpy.ops.armature.select_all(action='SELECT')
bpy.ops.armature.delete()
bpy.ops.object.mode_set(mode='OBJECT')
result = {"object": obj.name, "type": "ARMATURE"}
`, objName, objName, locStr)

		case "CURVE":
			curveType := getStringArg(args, "curve_type")
			if curveType == "" {
				curveType = "BEZIER"
			}
			code = fmt.Sprintf(`import bpy
curve = bpy.data.curves.new("%s_data", type='CURVE')
curve.dimensions = '3D'
curve.resolution_u = 12
spline = curve.splines.new("%s")
if "%s" == "BEZIER":
    spline.bezier_points.add(1)
    spline.bezier_points[0].co = (0, 0, 0)
    spline.bezier_points[0].handle_right_type = 'AUTO'
    spline.bezier_points[0].handle_left_type = 'AUTO'
    spline.bezier_points[1].co = (2, 0, 0)
    spline.bezier_points[1].handle_right_type = 'AUTO'
    spline.bezier_points[1].handle_left_type = 'AUTO'
else:
    spline.points.add(1)
    spline.points[0].co = (0, 0, 0, 1)
    spline.points[1].co = (2, 0, 0, 1)
obj = bpy.data.objects.new("%s", curve)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "CURVE", "curve_type": "%s"}
`, objName, curveType, curveType, objName, locStr, curveType)

		case "CURVES":
			code = fmt.Sprintf(`import bpy
curves = bpy.data.curves.new("%s_data", type='CURVES')
curves.surface_type = 'POLY'
obj = bpy.data.objects.new("%s", curves)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "CURVES"}
`, objName, objName, locStr)

		case "FONT":
			fontText := getStringArg(args, "font_text")
			if fontText == "" {
				fontText = "Text"
			}
			code = fmt.Sprintf(`import bpy
curve = bpy.data.curves.new("%s_data", type='FONT')
curve.body = "%s"
curve.size_x = 1.0
curve.size_y = 0.5
obj = bpy.data.objects.new("%s", curve)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "FONT", "text": "%s"}
`, objName, fontText, objName, locStr, fontText)

		case "GREASEPENCIL":
			code = fmt.Sprintf(`import bpy
gp = bpy.data.grease_pencils.new("%s_data")
layer = gp.layers.new("lines", set_active=True)
obj = bpy.data.objects.new("%s", gp)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "GREASEPENCIL", "layer": layer.name}
`, objName, objName, locStr)

		case "LATTICE":
			code = fmt.Sprintf(`import bpy
lattice = bpy.data.lattices.new("%s_data")
obj = bpy.data.objects.new("%s", lattice)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "LATTICE"}
`, objName, objName, locStr)

		case "LIGHT_PROBE":
			code = fmt.Sprintf(`import bpy
probe = bpy.data.light_probes.new("%s_data", type='PLANAR')
obj = bpy.data.objects.new("%s", probe)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "LIGHT_PROBE"}
`, objName, objName, locStr)

		case "META":
			code = fmt.Sprintf(`import bpy
mball = bpy.data.metaballs.new("%s_data")
mball.resolution = 0.05
ele = mball.elements.new()
ele.co = (0, 0, 0)
ele.radius = 1.0
obj = bpy.data.objects.new("%s", mball)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "META"}
`, objName, objName, locStr)

		case "POINTCLOUD":
			code = fmt.Sprintf(`import bpy
pc = bpy.data.pointclouds.new("%s_data")
obj = bpy.data.objects.new("%s", pc)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "POINTCLOUD"}
`, objName, objName, locStr)

		case "SPEAKER":
			code = fmt.Sprintf(`import bpy
speaker = bpy.data.speakers.new("%s_data")
obj = bpy.data.objects.new("%s", speaker)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "SPEAKER"}
`, objName, objName, locStr)

		case "SURFACE":
			code = fmt.Sprintf(`import bpy
surface = bpy.data.curves.new("%s_data", type='SURFACE')
surface.dimensions = '3D'
spline = surface.splines.new('NURBS')
spline.points.add(3)
for i, pt in enumerate(spline.points):
    pt.co = (float(i), 0, 0, 1)
obj = bpy.data.objects.new("%s", surface)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "SURFACE"}
`, objName, objName, locStr)

		case "VOLUME":
			code = fmt.Sprintf(`import bpy
volume = bpy.data.volumes.new("%s_data")
obj = bpy.data.objects.new("%s", volume)
obj.location = %s
bpy.context.scene.collection.objects.link(obj)
bpy.context.view_layer.objects.active = obj
obj.select_set(True)
result = {"object": obj.name, "type": "VOLUME"}
`, objName, objName, locStr)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown object_type: " + objType}
		}

		return execAndFormat(bc, code)
	}
}

func handleGreasePencilManage(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		objName := getStringArg(args, "object_name")
		action := getStringArg(args, "action")
		if objName == "" || action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name and action are required"}
		}

		switch action {
		case "list_layers":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layers = [{"name": l.name, "opacity": l.opacity, "visible": l.hide, "active": l == gp.layers.active} for l in gp.layers]
    result = {"object": obj.name, "layers": layers, "count": len(layers)}
`, objName)
			return execAndFormat(bc, code)

		case "add_layer":
			layerName := getStringArg(args, "layer_name")
			if layerName == "" {
				layerName = "NewLayer"
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.new("%s", set_active=True)
    result = {"added": layer.name, "total": len(gp.layers)}
`, objName, layerName)
			return execAndFormat(bc, code)

		case "remove_layer":
			layerName := getStringArg(args, "layer_name")
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.get("%s")
    if layer:
        gp.layers.remove(layer)
        result = {"removed": "%s", "remaining": len(gp.layers)}
    else:
        result = {"error": "layer not found"}
`, objName, layerName, layerName)
			return execAndFormat(bc, code)

		case "list_modifiers":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    mods = [{"name": m.name, "type": m.type, "show": m.show_viewport} for m in obj.grease_pencil_modifiers]
    result = {"object": obj.name, "modifiers": mods, "count": len(mods)}
`, objName)
			return execAndFormat(bc, code)

		case "add_modifier":
			modType := getStringArg(args, "modifier_type")
			if modType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "modifier_type is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    mod = obj.grease_pencil_modifiers.new("%s", type="%s")
    result = {"added": mod.name, "type": mod.type}
`, objName, modType, modType)
			return execAndFormat(bc, code)

		case "remove_modifier":
			modName := getStringArg(args, "modifier_name")
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    mod = obj.grease_pencil_modifiers.get("%s")
    if mod:
        obj.grease_pencil_modifiers.remove(mod)
        result = {"removed": "%s"}
    else:
        result = {"error": "modifier not found"}
`, objName, modName, modName)
			return execAndFormat(bc, code)

		case "draw_stroke":
			pointsRaw, ok := args["points"].([]any)
			if !ok || len(pointsRaw) == 0 {
				return nil, &Error{Code: CodeInvalidParams, Message: "points array is required"}
			}
			vertStrs := []string{}
			for _, p := range pointsRaw {
				if arr, ok := p.([]any); ok && len(arr) >= 3 {
					vertStrs = append(vertStrs, fmt.Sprintf(`(%.6f, %.6f, %.6f, 1.0)`, toFloat(arr[0]), toFloat(arr[1]), toFloat(arr[2])))
				}
			}
			pressure := 1.0
			if v, ok := args["pressure"].(float64); ok {
				pressure = v
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer is None:
        layer = gp.layers.new("lines", set_active=True)
    frame = layer.active_frame
    if frame is None:
        frame = layer.frames.new(bpy.context.scene.frame_current)
    stroke = frame.strokes.new()
    stroke.display_mode = '3DSPACE'
    stroke.draw_mode = 'LINE'
    points = [%s]
    stroke.points.add(len(points))
    for i, pt in enumerate(points):
        stroke.points[i].co = pt[:3]
        stroke.points[i].pressure = %.4f
        stroke.points[i].strength = 1.0
    result = {"stroke": "created", "points": len(points), "layer": layer.name}
`, objName, strings.Join(vertStrs, ", "), pressure)
			return execAndFormat(bc, code)

		case "fill":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer and layer.active_frame:
        stroke = layer.active_frame.strokes.active
        if stroke:
            stroke.draw_mode = 'SOLID'
            result = {"filled": True}
        else:
            result = {"error": "no active stroke"}
    else:
        result = {"error": "no active frame"}
`, objName)
			return execAndFormat(bc, code)

		case "extrude":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.extrude_move()
    result = {"extruded": True}
`, objName)
			return execAndFormat(bc, code)

		case "delete":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.delete()
    result = {"deleted": True}
`, objName)
			return execAndFormat(bc, code)

		case "dissolve":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.dissolve()
    result = {"dissolved": True}
`, objName)
			return execAndFormat(bc, code)

		case "duplicate":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.duplicate_move()
    result = {"duplicated": True}
`, objName)
			return execAndFormat(bc, code)

		case "clean_loose":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.clean_loose()
    result = {"cleaned": True}
`, objName)
			return execAndFormat(bc, code)

		case "convert_curve_type":
			ct := getStringArg(args, "curve_type")
			if ct == "" {
				ct = "BEZIER"
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.convert_curve_type(type="%s")
    result = {"converted_to": "%s"}
`, objName, ct, ct)
			return execAndFormat(bc, code)

		case "set_cyclic":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.cyclical_set()
    result = {"cyclic_set": True}
`, objName)
			return execAndFormat(bc, code)

		case "list_frames":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer is None:
        result = {"error": "no active layer"}
    else:
        frames = [{"frame": f.frame_number, "strokes": len(f.strokes)} for f in layer.frames]
        result = {"layer": layer.name, "frames": frames, "count": len(frames)}
`, objName)
			return execAndFormat(bc, code)

		case "add_frame":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer is None:
        result = {"error": "no active layer"}
    else:
        frame = layer.frames.new(bpy.context.scene.frame_current)
        result = {"added_frame": frame.frame_number}
`, objName)
			return execAndFormat(bc, code)

		case "remove_frame":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer and layer.active_frame:
        layer.frames.remove(layer.active_frame)
        result = {"removed_frame": True}
    else:
        result = {"error": "no active frame"}
`, objName)
			return execAndFormat(bc, code)

		case "active_frame_delete":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    gp = obj.data
    layer = gp.layers.active
    if layer and layer.active_frame:
        layer.frames.remove(layer.active_frame)
        result = {"deleted_active_frame": True}
    else:
        result = {"error": "no active frame"}
`, objName)
			return execAndFormat(bc, code)

		case "delete_frame":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.delete_frame()
    result = {"deleted_frame": True}
`, objName)
			return execAndFormat(bc, code)

		case "brush_stroke":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.brush_stroke()
    result = {"brush_stroke": True}
`, objName)
			return execAndFormat(bc, code)

		case "caps_set":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.caps_set(type="ROUND")
    result = {"caps_set": "ROUND"}
`, objName)
			return execAndFormat(bc, code)

		case "copy":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.copy()
    result = {"copied": True}
`, objName)
			return execAndFormat(bc, code)

		case "delete_breakdown":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.delete_breakdown()
    result = {"deleted_breakdown": True}
`, objName)
			return execAndFormat(bc, code)

		case "erase_box":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.erase(type="BOX")
    result = {"erased_box": True}
`, objName)
			return execAndFormat(bc, code)

		case "erase_lasso":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.erase(type="LASSO")
    result = {"erased_lasso": True}
`, objName)
			return execAndFormat(bc, code)

		case "frame_clean_duplicate":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.frame_clean_duplicate()
    result = {"frame_clean_duplicate": True}
`, objName)
			return execAndFormat(bc, code)

		case "bake_grease_pencil_animation":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None or obj.type not in ('GPENCIL', 'GREASEPENCIL'):
    result = {"error": "not a Grease Pencil object"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.gpencil.bake_animation()
    result = {"baked_animation": True}
`, objName)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleCompositorNodes(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		action := getStringArg(args, "action")
		if action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "action is required"}
		}

		switch action {
		case "list_types":
			code := `import bpy
# Enumerate all available compositor node types
types = set()
for attr in dir(bpy.types):
    obj = getattr(bpy.types, attr, None)
    try:
        if hasattr(obj, '__bases__') and any('CompositorNode' in str(b) for b in obj.__bases__ if hasattr(b, '__name__')):
            types.add(attr)
    except:
        pass
# Fallback: known Blender compositor node types (5.x)
known = [
    "CompositorNodeAlphaOver", "CompositorNodeAntiAliasing", "CompositorNodeBilateralblur",
    "CompositorNodeBlackbody", "CompositorNodeBokehBlur", "CompositorNodeBokehImage",
    "CompositorNodeBoxMask", "CompositorNodeBrightContrast", "CompositorNodeCameraData",
    "CompositorNodeChannelMatte", "CompositorNodeChromaMatte", "CompositorNodeColorBalance",
    "CompositorNodeColorCorrection", "CompositorNodeColorMatte", "CompositorNodeColorRamp",
    "CompositorNodeColorSpill", "CompositorNodeCombineHSVA", "CompositorNodeCombineRGBA",
    "CompositorNodeCombineYUVA", "CompositorNodeCombineYZA", "CompositorNodeConvertColorSpace",
    "CompositorNodeCrop", "CompositorNodeCryptomatte", "CompositorNodeCryptomatteV2",
    "CompositorNodeCurveRGB", "CompositorNodeCurveVec", "CompositorNodeCustomGroup",
    "CompositorNodeDBlur", "CompositorNodeDefocus", "CompositorNodeDenoise",
    "CompositorNodeDespeckle", "CompositorNodeDiffMatte", "CompositorNodeDilateErode",
    "CompositorNodeDisplace", "CompositorNodeDistanceMatte", "CompositorNodeDoubleEdgeMask",
    "CompositorNodeEllipseMask", "CompositorNodeExposure", "CompositorNodeFilter",
    "CompositorNodeFlip", "CompositorNodeFunction", "CompositorNodeGamma",
    "CompositorNodeGlare", "CompositorNodeGroup", "CompositorNodeHueCorrect",
    "CompositorNodeHueSat", "CompositorNodeIDMask", "CompositorNodeImage",
    "CompositorNodeInpaint", "CompositorNodeInvert", "CompositorNodeKeying",
    "CompositorNodeKeyingScreen", "CompositorNodeKuwahara", "CompositorNodeLensdist",
    "CompositorNodeLevels", "CompositorNodeLumaMatte", "CompositorNodeMapUV",
    "CompositorNodeMapValue", "CompositorNodeMask", "CompositorNodeMath",
    "CompositorNodeMixRGB", "CompositorNodeMovieClip", "CompositorNodeMovieDistortion",
    "CompositorNodeNormal", "CompositorNodeNormalize", "CompositorNodeOutputFile",
    "CompositorNodePixelate", "CompositorNodePlaneTrackDeform", "CompositorNodePosterize",
    "CompositorNodePremulKey", "CompositorNodeRGB", "CompositorNodeRGBToBW",
    "CompositorNodeRLayers", "CompositorNodeRotate", "CompositorNodeScale",
    "CompositorNodeSceneTime", "CompositorNodeSeparateHSVA", "CompositorNodeSeparateRGBA",
    "CompositorNodeSeparateYUVA", "CompositorNodeSeparateYZA", "CompositorNodeSetAlpha",
    "CompositorNodeSplit", "CompositorNodeStabilize", "CompositorNodeSwitch",
    "CompositorNodeSwitchView", "CompositorNodeTexture", "CompositorNodeTime",
    "CompositorNodeTonemap", "CompositorNodeTrackPos", "CompositorNodeTransform",
    "CompositorNodeTranslate", "CompositorNodeTree", "CompositorNodeVecBlur",
    "CompositorNodeViewer", "CompositorNodeZcombine",
]
all_types = sorted(types | set(known))
result = {"compositor_node_types": all_types, "count": len(all_types)}
`
			return execAndFormat(bc, code)

		case "list_nodes":
			code := `import bpy
scene = bpy.context.scene
if not scene.use_nodes:
    scene.use_nodes = True
nodes = [{"name": n.name, "type": n.type, "label": n.label} for n in scene.node_tree.nodes]
result = {"nodes": nodes, "count": len(nodes), "use_nodes": scene.use_nodes}
`
			return execAndFormat(bc, code)

		case "add_node":
			nodeType := getStringArg(args, "node_type")
			if nodeType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "node_type is required"}
			}
			nodeName := getStringArg(args, "node_name")
			if nodeName == "" {
				nodeName = nodeType
			}
			code := fmt.Sprintf(`import bpy
scene = bpy.context.scene
if not scene.use_nodes:
    scene.use_nodes = True
node = scene.node_tree.nodes.new(type="%s")
node.label = "%s"
result = {"added": node.name, "type": node.type, "label": node.label}
`, nodeType, nodeName)
			return execAndFormat(bc, code)

		case "remove_node":
			nodeName := getStringArg(args, "node_name")
			code := fmt.Sprintf(`import bpy
scene = bpy.context.scene
node = scene.node_tree.nodes.get("%s")
if node:
    scene.node_tree.nodes.remove(node)
    result = {"removed": "%s"}
else:
    result = {"error": "node not found"}
`, nodeName, nodeName)
			return execAndFormat(bc, code)

		case "connect_nodes":
			fromNode := getStringArg(args, "from_node")
			fromSocket := getStringArg(args, "from_socket")
			toNode := getStringArg(args, "to_node")
			toSocket := getStringArg(args, "to_socket")
			if fromNode == "" || fromSocket == "" || toNode == "" || toSocket == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "from_node, from_socket, to_node, to_socket are all required"}
			}
			code := fmt.Sprintf(`import bpy
scene = bpy.context.scene
tree = scene.node_tree
src = tree.nodes.get("%s")
dst = tree.nodes.get("%s")
if src is None or dst is None:
    result = {"error": "node not found"}
else:
    src_out = src.outputs.get("%s")
    dst_in = dst.inputs.get("%s")
    if src_out is None or dst_in is None:
        result = {"error": "socket not found"}
    else:
        tree.links.new(src_out, dst_in)
        result = {"connected": "%s.%s -> %s.%s"}
`, fromNode, toNode, fromSocket, toSocket, fromNode, fromSocket, toNode, toSocket)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleManageShaderNodes(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		action := getStringArg(args, "action")
		if action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "action is required"}
		}
		objName := getStringArg(args, "object_name")
		if objName == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
		}

		switch action {
		case "list_types":
			code := `import bpy
types = [cls.__name__ for cls in bpy.types.ShaderNode.__subclasses__()]
types.sort()
result = {"shader_node_types": types, "count": len(types)}
`
			return execAndFormat(bc, code)

		case "list_nodes":
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mat = obj.active_material
    if mat is None:
        result = {"error": "no active material"}
    elif not mat.use_nodes:
        result = {"error": "material does not use nodes"}
    else:
        nodes = [{"name": n.name, "type": n.type, "label": n.label} for n in mat.node_tree.nodes]
        result = {"material": mat.name, "nodes": nodes, "count": len(nodes)}
`, objName)
			return execAndFormat(bc, code)

		case "add_node":
			nodeType := getStringArg(args, "node_type")
			if nodeType == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "node_type is required"}
			}
			nodeName := getStringArg(args, "node_name")
			if nodeName == "" {
				nodeName = nodeType
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mat = obj.active_material
    if mat is None:
        result = {"error": "no active material"}
    else:
        if not mat.use_nodes:
            mat.use_nodes = True
        node = mat.node_tree.nodes.new(type="%s")
        node.label = "%s"
        result = {"added": node.name, "type": node.type, "label": node.label}
`, objName, nodeType, nodeName)
			return execAndFormat(bc, code)

		case "remove_node":
			nodeName := getStringArg(args, "node_name")
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mat = obj.active_material
    if mat is None:
        result = {"error": "no active material"}
    else:
        node = mat.node_tree.nodes.get("%s")
        if node:
            mat.node_tree.nodes.remove(node)
            result = {"removed": "%s"}
        else:
            result = {"error": "node not found"}
`, objName, nodeName, nodeName)
			return execAndFormat(bc, code)

		case "connect_nodes":
			fromNode := getStringArg(args, "from_node")
			fromSocket := getStringArg(args, "from_socket")
			toNode := getStringArg(args, "to_node")
			toSocket := getStringArg(args, "to_socket")
			if fromNode == "" || fromSocket == "" || toNode == "" || toSocket == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "from_node, from_socket, to_node, to_socket are all required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj is None:
    result = {"error": "object not found"}
else:
    mat = obj.active_material
    if mat is None:
        result = {"error": "no active material"}
    else:
        tree = mat.node_tree
        src = tree.nodes.get("%s")
        dst = tree.nodes.get("%s")
        if src is None or dst is None:
            result = {"error": "node not found"}
        else:
            src_out = src.outputs.get("%s")
            dst_in = dst.inputs.get("%s")
            if src_out is None or dst_in is None:
                result = {"error": "socket not found"}
            else:
                tree.links.new(src_out, dst_in)
                result = {"connected": "%s.%s -> %s.%s"}
`, objName, fromNode, toNode, fromSocket, toSocket, fromNode, fromSocket, toNode, toSocket)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleNodeWranglerOps(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		action := getStringArg(args, "action")
		if action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "action is required"}
		}
		objName := getStringArg(args, "object_name")
		treeType := getStringArg(args, "node_tree_type")
		if treeType == "" {
			treeType = "MATERIAL"
		}

		switch action {
		case "preview_link":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_preview_link()
        result = {"preview_link": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "swap_nodes":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_swap_nodes()
        result = {"swapped": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "mix_nodes":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_mix_nodes()
        result = {"mixed": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "collapse_all":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        for node in mat.node_tree.nodes:
            node.mute = True
        result = {"collapsed": len(mat.node_tree.nodes)}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "expand_all":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        for node in mat.node_tree.nodes:
            node.mute = False
        result = {"expanded": len(mat.node_tree.nodes)}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "frame_selected":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_frame_selected()
        result = {"framed": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "add_texture_setup":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_add_texture(setup=True)
        result = {"texture_setup": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "connect_viewer":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_viewer()
        result = {"viewer_connected": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "disconnect_viewer":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_remove_viewer()
        result = {"viewer_disconnected": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		case "node_switch":
			if objName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "object_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get("%s")
if obj and obj.type == 'MESH' and obj.data.materials:
    mat = obj.data.materials[0]
    if mat.use_nodes:
        bpy.context.view_layer.objects.active = obj
        bpy.ops.node.nw_node_switch()
        result = {"switched": True}
    else:
        result = {"error": "material has no nodes"}
else:
    result = {"error": "object not found or no material"}
`, objName)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handlePoseLibraryOps(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		action := getStringArg(args, "action")
		if action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "action is required"}
		}
		armName := getStringArg(args, "armature_name")

		switch action {
		case "list_poses":
			if armName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "armature_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get('%s')
if obj is None or obj.type not in ('ARMATURE', 'ARMATURE_DATA'):
    result = {"error": "not an armature"}
else:
    actions = []
    for a in bpy.data.actions:
        has_pose = False
        fc_list = []
        if a.is_action_legacy:
            fc_list = list(a.fcurves)
        else:
            for layer in a.layers:
                for strip in layer.strips:
                    for cb in strip.channelbags:
                        fc_list.extend(list(cb.fcurves))
        for fc in fc_list:
            if fc.data_path.startswith("pose.bones"):
                has_pose = True
                break
        if has_pose:
            actions.append({"name": a.name, "frame_range": list(a.frame_range), "users": a.users})
    result = {"armature": obj.name, "actions": actions, "count": len(actions)}
`, escapePyStr(armName))
			return execAndFormat(bc, code)

		case "save_pose":
			poseName := getStringArg(args, "pose_name")
			if armName == "" || poseName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "armature_name and pose_name are required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get('%s')
if obj is None or obj.type not in ('ARMATURE', 'ARMATURE_DATA'):
    result = {"error": "not an armature"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    action = bpy.data.actions.new(name='%s')
    action.use_fake_user = True
    stored = 0
    for pb in obj.pose.bones:
        bname = pb.name
        if pb.bone.use_location:
            for i in range(3):
                dp = "pose.bones['" + bname + "'].location"
                fc = action.fcurves.new(data_path=dp, index=i)
                fc.keyframe_points.add(1)
                fc.keyframe_points[0].co = (1, pb.location[i])
            stored += 1
        if pb.bone.use_rotation:
            for i in range(3):
                dp = "pose.bones['" + bname + "'].rotation_euler"
                fc = action.fcurves.new(data_path=dp, index=i)
                fc.keyframe_points.add(1)
                fc.keyframe_points[0].co = (1, pb.rotation_euler[i])
            stored += 1
        if pb.bone.use_scale:
            for i in range(3):
                dp = "pose.bones['" + bname + "'].scale"
                fc = action.fcurves.new(data_path=dp, index=i)
                fc.keyframe_points.add(1)
                fc.keyframe_points[0].co = (1, pb.scale[i])
            stored += 1
    result = {"saved": '%s', "action": action.name, "bones_stored": stored, "armature": obj.name}
`, escapePyStr(armName), escapePyStr(poseName), escapePyStr(poseName))
			return execAndFormat(bc, code)

		case "apply_pose":
			poseName := getStringArg(args, "pose_name")
			if armName == "" || poseName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "armature_name and pose_name are required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get('%s')
if obj is None or obj.type not in ('ARMATURE', 'ARMATURE_DATA'):
    result = {"error": "not an armature"}
else:
    action = bpy.data.actions.get('%s')
    if action is None:
        result = {"error": "action not found: %s"}
    else:
        bpy.context.view_layer.objects.active = obj
        obj.select_set(True)
        fc_list = []
        if action.is_action_legacy:
            fc_list = list(action.fcurves)
        else:
            for layer in action.layers:
                for strip in layer.strips:
                    for cb in strip.channelbags:
                        fc_list.extend(list(cb.fcurves))
        applied = 0
        for fc in fc_list:
            path = fc.data_path
            if "pose.bones" in path:
                bone_name = path.split("pose.bones['")[1].split("']")[0]
                pb = obj.pose.bones.get(bone_name)
                if pb:
                    if "location" in path:
                        pb.location[fc.array_index] = fc.evaluate(1)
                        applied += 1
                    elif "rotation_euler" in path:
                        pb.rotation_euler[fc.array_index] = fc.evaluate(1)
                        applied += 1
                    elif "scale" in path:
                        pb.scale[fc.array_index] = fc.evaluate(1)
                        applied += 1
        result = {"applied": '%s', "armature": obj.name, "bones_modified": applied}
`, escapePyStr(armName), escapePyStr(poseName), escapePyStr(poseName), escapePyStr(poseName))
			return execAndFormat(bc, code)

		case "create_library":
			libPath := getStringArg(args, "library_path")
			if libPath == "" {
				libPath = "//pose_library"
			}
			code := fmt.Sprintf(`import bpy, os
# Create asset library directory structure
lib_path = r"%s"
os.makedirs(lib_path, exist_ok=True)
# Create a .blend file for pose library
blend_path = os.path.join(lib_path, "poses.blend")
bpy.ops.wm.save_as_mainfile(filepath=blend_path, check_existing=False)
result = {"library_path": lib_path, "blend_file": blend_path, "note": "Pose library created. Add this path as an Asset Library in Blender Preferences > File Paths."}
`, libPath)
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
	}
}

func handleRigifyOps(bc *blender.Client) ToolHandler {
	return func(args map[string]any) (any, *Error) {
		action := getStringArg(args, "action")
		if action == "" {
			return nil, &Error{Code: CodeInvalidParams, Message: "action is required"}
		}

		switch action {
		case "generate_rig":
			armName := getStringArg(args, "armature_name")
			if armName == "" {
				return nil, &Error{Code: CodeInvalidParams, Message: "armature_name is required"}
			}
			code := fmt.Sprintf(`import bpy
obj = bpy.data.objects.get('%s')
if obj is None or obj.type not in ('ARMATURE', 'ARMATURE_DATA'):
    result = {"error": "not an armature"}
else:
    bpy.context.view_layer.objects.active = obj
    obj.select_set(True)
    bpy.ops.pose.rigify_generate()
    result = {"generated": obj.name}
`, armName)
			return execAndFormat(bc, code)

		case "add_metarig":
			metarigType := getStringArg(args, "metarig_type")
			if metarigType == "" {
				metarigType = "human"
			}
			// Map metarig_type to actual bpy.ops operator names
			opMap := map[string]string{
				"basic":          "bpy.ops.object.armature_basic_metarig_add(",
				"human":          "bpy.ops.object.armature_human_metarig_add(",
				"quadruped":      "bpy.ops.object.animal_metarig_add(",
				"bird":           "bpy.ops.object.animal_metarig_add(",
				"cat":            "bpy.ops.object.animal_metarig_add(",
				"horse":          "bpy.ops.object.animal_metarig_add(",
				"monkey":         "bpy.ops.object.animal_metarig_add(",
				"shark":          "bpy.ops.object.animal_metarig_add(",
				"wolf":           "bpy.ops.object.animal_metarig_add(",
				"arm.finger":     "bpy.ops.object.armature_human_metarig_add(",
				"leg.plane.02":   "bpy.ops.object.armature_human_metarig_add(",
				"spine.basic.01": "bpy.ops.object.armature_human_metarig_add(",
				"spine.reptile.01": "bpy.ops.object.armature_human_metarig_add(",
				"head.basic.01":  "bpy.ops.object.armature_human_metarig_add(",
			}
			opCall, ok := opMap[metarigType]
			if !ok {
				// Default to human for unknown types
				opCall = "bpy.ops.object.armature_human_metarig_add("
			}
			code := fmt.Sprintf(`import bpy
%senter_editmode=True, location=(0, 0, 0))
obj = bpy.context.active_object
if obj:
    result = {"added": obj.name, "type": "%s"}
else:
    result = {"error": "failed to add metarig"}
`, opCall, metarigType)
			return execAndFormat(bc, code)

		case "list_metarigs":
			code := `import bpy
metarigs = ["basic", "human", "quadruped", "bird", "cat", "horse", "monkey", "shark", "wolf", "arm.finger", "leg.plane.02", "spine.basic.01", "spine.reptile.01", "head.basic.01"]
result = {"metarigs": metarigs, "count": len(metarigs)}
`
			return execAndFormat(bc, code)

		case "list_rig_types":
			code := `import bpy
try:
    rig_types = [r.identifier for r in bpy.types.RigType.__subclasses__()]
    result = {"rig_types": rig_types, "count": len(rig_types)}
except AttributeError:
    result = {"error": "Rigify addon not installed or not compatible with this Blender version", "hint": "Enable Rigify in Edit > Preferences > Add-ons"}
`
			return execAndFormat(bc, code)

		default:
			return nil, &Error{Code: CodeInvalidParams, Message: "unknown action: " + action}
		}
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
