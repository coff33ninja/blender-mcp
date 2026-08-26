<#
.SYNOPSIS
    Probes Blender's MCP-exposed capabilities and logs results to markdown.
.DESCRIPTION
    Connects to the Blender MCP server, runs probe queries across all major
    capability categories, and writes a structured markdown report.
    Run this after Blender version updates or add-on changes.
.NOTES
    Requires: Blender running with MCP add-on on localhost:9876
    Output:   probe-results.md in the script's directory
#>

param(
    [string]$McpPath = (Join-Path $PSScriptRoot "blender-mcp.exe"),
    [string]$OutputPath = ""
)

if (-not $OutputPath) {
    $OutputPath = Join-Path $PSScriptRoot "probe-results.md"
}

# ── helpers ──────────────────────────────────────────────────────────────────

function Send-Rpc {
    param([int]$Id, [string]$Method, [hashtable]$Params, [System.Diagnostics.Process]$Proc)
    $msg = @{ jsonrpc = "2.0"; id = $Id; method = $Method; params = $Params }
    $json = $msg | ConvertTo-Json -Depth 10 -Compress
    $Proc.StandardInput.WriteLine($json)
    Start-Sleep -Milliseconds 300
}

function Invoke-McpProbe {
    param(
        [int]$Id,
        [string]$Description,
        [string]$Code,
        [System.Diagnostics.Process]$Proc
    )
    Send-Rpc -Id $Id -Method "tools/call" -Params @{ name = "execute_code"; arguments = @{ code = $Code } } -Proc $Proc
    return $Id
}

function Start-McpServer {
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $McpPath
    $psi.UseShellExecute = $false
    $psi.RedirectStandardInput = $true
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.WorkingDirectory = Split-Path $McpPath
    $proc = [System.Diagnostics.Process]::Start($psi)

    Send-Rpc -Id 1 -Method "initialize" -Params @{
        protocolVersion = "2025-03-26"
        capabilities    = @{}
        clientInfo      = @{ name = "blender-probe"; version = "1.0" }
    } -Proc $proc
    Send-Rpc -Id 0 -Method "notifications/initialized" -Params @{} -Proc $proc

    return $proc
}

function Read-McpResults {
    param([System.Diagnostics.Process]$Proc, [int]$TimeoutSec = 12)
    $proc.StandardInput.Close()

    # Read stdout in background job with timeout
    $job = Start-Job -ScriptBlock {
        param($p)
        $lines = @()
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        while ($sw.Elapsed.TotalSeconds -lt 30) {
            if ($p.StandardOutput.EndOfStream) { break }
            $line = $p.StandardOutput.ReadLine()
            if ($line) { $lines += $line }
        }
        return ($lines -join "`n")
    } -ArgumentList $proc

    $completed = Wait-Job $job -Timeout $TimeoutSec
    if (-not $completed) {
        Write-Host "  Read timeout after ${TimeoutSec}s, collecting partial..." -ForegroundColor Yellow
        Stop-Job $job 2>$null
    }
    $output = Receive-Job $job
    Remove-Job $job -Force 2>$null

    if (-not $proc.HasExited) {
        try { $proc.Kill() } catch {}
        $proc.WaitForExit(2000) | Out-Null
    }

    $results = @{}
    foreach ($line in ($output -split "`n")) {
        $line = $line.Trim()
        if (-not $line) { continue }
        try {
            $json = $line | ConvertFrom-Json
            if ($json.id -is [int] -and $json.id -ge 10 -and $json.result.content) {
                $text = $json.result.content[0].text
                $results[$json.id] = $text | ConvertFrom-Json
            }
        } catch {}
    }
    return $results
}

# ── probe definitions ────────────────────────────────────────────────────────

$probes = @(
    @{
        Id   = 10
        Desc = "Scene Overview"
        Code = @'
import bpy
scene = bpy.context.scene
result = {
    "name": scene.name,
    "render_engine": scene.render.engine,
    "resolution_x": scene.render.resolution_x,
    "resolution_y": scene.render.resolution_y,
    "resolution_percentage": scene.render.resolution_percentage,
    "fps": scene.render.fps,
    "frame_start": scene.frame_start,
    "frame_end": scene.frame_end,
    "frame_current": scene.frame_current,
    "file_format": scene.render.image_settings.file_format,
    "color_mode": scene.render.image_settings.color_mode,
    "filepath": scene.filepath,
    "world": scene.world.name if scene.world else None,
    "use_nodes": scene.use_nodes,
}
'@
    },
    @{
        Id   = 11
        Desc = "Object Types"
        Code = @'
import bpy
types = sorted([t.identifier for t in bpy.types.Object.bl_rna.properties['type'].enum_items_static])
result = {"object_types": types, "count": len(types)}
'@
    },
    @{
        Id   = 12
        Desc = "Render Engines"
        Code = @'
import bpy
engines = sorted([e.identifier for e in bpy.types.RenderEngine.__subclasses__()])
result = {"engines": engines, "current": bpy.context.scene.render.engine}
'@
    },
    @{
        Id   = 13
        Desc = "Modifier Types"
        Code = @'
import bpy
bpy.ops.mesh.primitive_cube_add()
obj = bpy.context.active_object
types = sorted([item.identifier for item in bpy.types.Modifier.bl_rna.properties['type'].enum_items_static])
bpy.data.objects.remove(obj)
result = {"modifier_types": types, "count": len(types)}
'@
    },
    @{
        Id   = 14
        Desc = "Constraint Types"
        Code = @'
import bpy
bpy.ops.mesh.primitive_cube_add()
obj = bpy.context.active_object
const_enum = bpy.types.ObjectConstraint.bl_rna.properties.get('type')
if const_enum and hasattr(const_enum, 'enum_items_static'):
    types = sorted([item.identifier for item in const_enum.enum_items_static])
else:
    types = []
bpy.data.objects.remove(obj)
result = {"constraint_types": types, "count": len(types)}
'@
    },
    @{
        Id   = 15
        Desc = "Physics Categories"
        Code = @'
import bpy
bpy.ops.mesh.primitive_cube_add()
obj = bpy.context.active_object
physics = {}
# Rigid body
bpy.ops.rigidbody.object_add()
physics["rigid_body"] = {"type": obj.rigid_body.type, "mass": obj.rigid_body.mass}
bpy.ops.rigidbody.object_remove()
# Cloth
bpy.ops.object.modifier_add(type='CLOTH')
physics["cloth"] = {"enabled": True}
obj.modifiers.remove(obj.modifiers[-1])
# Fluid
bpy.ops.object.modifier_add(type='FLUID')
physics["fluid"] = {"types": ["DOMAIN", "FLOW", "EFFECTOR"]}
obj.modifiers.remove(obj.modifiers[-1])
# Soft body
bpy.ops.object.modifier_add(type='SOFT_BODY')
physics["soft_body"] = {"enabled": True}
obj.modifiers.remove(obj.modifiers[-1])
# Force field
obj.field.type = 'FORCE'
physics["force_field"] = {"types": sorted([t.identifier for t in bpy.types.Field.bl_rna.properties['type'].enum_items_static])}
obj.field.type = 'NONE'
# Particle system
bpy.ops.object.particle_system_add()
physics["particle_system"] = {"enabled": True}
obj.particle_systems.clear()
bpy.data.objects.remove(obj)
result = {"physics": physics}
'@
    },
    @{
        Id   = 16
        Desc = "File Formats"
        Code = @'
import bpy
img = bpy.context.scene.render.image_settings
formats = sorted([item.identifier for item in img.bl_rna.properties['file_format'].enum_items_static])
color_modes = sorted([item.identifier for item in img.bl_rna.properties['color_mode'].enum_items_static])
result = {"file_formats": formats, "color_modes": color_modes}
'@
    },
    @{
        Id   = 17
        Desc = "Installed Add-ons"
        Code = @'
import bpy
addons = []
for a in bpy.context.preferences.addons:
    info = bpy.utils.addon_info.get_addon_info(a.module)
    addons.append({
        "module": a.module,
        "name": info.name if info else a.module,
        "version": ".".join(str(v) for v in info.version) if info and info.version else "unknown",
    })
addons.sort(key=lambda x: x["module"])
result = {"addons": addons, "count": len(addons)}
'@
    },
    @{
        Id   = 18
        Desc = "Shader Node Classes"
        Code = @'
import bpy
classes = sorted([n.__name__ for n in bpy.types.Node.__subclasses__()])
result = {"node_classes": classes, "count": len(classes)}
'@
    },
    @{
        Id   = 19
        Desc = "Render Pass Toggles"
        Code = @'
import bpy
vl = bpy.context.view_layer
passes = {k: getattr(vl, k) for k in dir(vl) if k.startswith('use_pass_')}
result = {"view_layer": vl.name, "passes": passes}
'@
    },
    @{
        Id   = 20
        Desc = "World Environment"
        Code = @'
import bpy
world = bpy.context.scene.world
if world and world.use_nodes and world.node_tree:
    nodes = [{"name": n.name, "type": n.type} for n in world.node_tree.nodes]
    result = {"world": world.name, "use_nodes": True, "nodes": nodes}
elif world:
    result = {"world": world.name, "use_nodes": False, "nodes": []}
else:
    result = {"world": None}
'@
    },
    @{
        Id   = 21
        Desc = "Image/Texture Node Types"
        Code = @'
import bpy
# Enumerate all available texture node types for shader editor
shader_nodes = [n.__name__ for n in bpy.types.ShaderNode.__subclasses__()]
compositor_nodes = [n.__name__ for n in bpy.types.CompositorNode.__subclasses__()]
geometry_nodes = [n.__name__ for n in bpy.types.GeometryNode.__subclasses__()] if hasattr(bpy.types, 'GeometryNode') else []
result = {
    "shader_nodes": sorted(shader_nodes),
    "shader_node_count": len(shader_nodes),
    "compositor_nodes": sorted(compositor_nodes),
    "compositor_node_count": len(compositor_nodes),
    "geometry_nodes": sorted(geometry_nodes),
    "geometry_node_count": len(geometry_nodes),
}
'@
    },
    @{
        Id   = 22
        Desc = "Grease Pencil Capabilities"
        Code = @'
import bpy
# Just check if GP types exist, don't enumerate everything
gp_ops = [n for n in dir(bpy.ops.grease_pencil) if not n.startswith('_')] if hasattr(bpy.ops, 'grease_pencil') else []
gpencil_mods = [t.identifier for t in bpy.types.Modifier.bl_rna.properties['type'].enum_items_static if 'GPENCIL' in t.identifier or 'GREASE' in t.identifier]
result = {"gpencil_ops_count": len(gp_ops), "gpencil_ops_sample": gp_ops[:15], "gpencil_modifiers": gpencil_mods}
'@
    },
    @{
        Id   = 23
        Desc = "Brush Types"
        Code = @'
import bpy
brush_list = [{"name": b.name} for b in bpy.data.brushes]
result = {"brushes": brush_list, "count": len(brush_list)}
'@
    },
    @{
        Id   = 24
        Desc = "Collections & Scene Structure"
        Code = @'
import bpy
scene = bpy.context.scene
collections = []
for col in bpy.data.collections:
    collections.append({"name": col.name, "object_count": len(col.objects), "children": len(col.children)})
scene_objects = []
for obj in scene.objects:
    scene_objects.append({"name": obj.name, "type": obj.type, "visible": obj.visible_get()})
result = {
    "collections": collections,
    "collection_count": len(collections),
    "scene_objects": scene_objects,
    "object_count": len(scene_objects),
}
'@
    },
    @{
        Id   = 25
        Desc = "Operator Namespaces"
        Code = @'
import bpy
namespaces = sorted([n for n in dir(bpy.ops) if not n.startswith('_')])
result = {"operator_namespaces": namespaces, "count": len(namespaces)}
'@
    }
)

# ── run probes ───────────────────────────────────────────────────────────────

Write-Host "Starting MCP server..." -ForegroundColor Yellow
$proc = Start-McpServer

Write-Host "Running $($probes.Count) probes..." -ForegroundColor Yellow
foreach ($p in $probes) {
    Write-Host "  Sending [$($p.Id)] $($p.Desc)..." -ForegroundColor Gray -NoNewline
    Invoke-McpProbe -Id $p.Id -Description $p.Desc -Code $p.Code -Proc $proc
    Write-Host " sent" -ForegroundColor DarkGray
}

Write-Host "Collecting results..." -ForegroundColor Yellow
$results = Read-McpResults -Proc $proc

# ── write markdown ───────────────────────────────────────────────────────────

$sb = [System.Text.StringBuilder]::new()
[void]$sb.AppendLine("# Blender MCP Capability Probe")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("**Date:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')")
[void]$sb.AppendLine("**Blender MCP Server:** $McpPath")
[void]$sb.AppendLine("")

foreach ($p in $probes) {
    $data = $results[$p.Id]
    [void]$sb.AppendLine("## $($p.Desc)")
    [void]$sb.AppendLine("")

    if ($data) {
        $json = $data | ConvertTo-Json -Depth 10
        $codeFence = [char]0x60 + [char]0x60 + [char]0x60
        [void]$sb.AppendLine("${codeFence}json")
        [void]$sb.AppendLine($json)
        [void]$sb.AppendLine($codeFence)
    } else {
        $msg = [char]0x5F + "No result returned." + [char]0x5F
        [void]$sb.AppendLine($msg)
    }
    [void]$sb.AppendLine("")
}

# ── summary ──────────────────────────────────────────────────────────────────

[void]$sb.AppendLine("## Summary Table")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("| Category | Count |")
[void]$sb.AppendLine("|----------|-------|")

$summaryMap = @{
    11 = "Object Types"
    13 = "Modifier Types"
    14 = "Constraint Types"
    16 = "File Formats"
    17 = "Add-ons"
    18 = "Node Classes"
    21 = "Shader Nodes"
    22 = "Compositor Nodes"
    23 = "Geometry Nodes"
    25 = "Operators"
}

foreach ($kv in $summaryMap.GetEnumerator()) {
    $data = $results[$kv.Key]
    if ($data) {
        $count = ($data.PSObject.Properties | Where-Object { $_.Name -like "*count*" }).Value
        if (-not $count) {
            # Try to count array properties
            foreach ($prop in $data.PSObject.Properties) {
                if ($prop.Value -is [array]) {
                    $count = $prop.Value.Count
                    break
                }
            }
        }
        [void]$sb.AppendLine("| $($kv.Value) | $($count ?? 'N/A') |")
    }
}

[void]$sb.AppendLine("")
[void]$sb.AppendLine("---")
[void]$sb.AppendLine("*Generated by probe-blender.ps1*")

$sb.ToString() | Out-File -FilePath $OutputPath -Encoding UTF8
Write-Host ""
Write-Host "Done! Results written to: $OutputPath" -ForegroundColor Green
Write-Host "Open it and compare against current tools to find gaps." -ForegroundColor Cyan
