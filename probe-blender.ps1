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

# ── probe definitions ────────────────────────────────────────────────────────

$probes = @(
    @{
        Id   = 10
        Desc = "Scene Overview"
        Code = @'
import bpy
scene = bpy.context.scene
r = scene.render
result = {
    "name": scene.name,
    "render_engine": r.engine,
    "resolution_x": r.resolution_x,
    "resolution_y": r.resolution_y,
    "resolution_percentage": r.resolution_percentage,
    "fps": r.fps,
    "frame_start": scene.frame_start,
    "frame_end": scene.frame_end,
    "frame_current": scene.frame_current,
    "file_format": r.image_settings.file_format,
    "color_mode": r.image_settings.color_mode,
    "filepath": r.filepath,
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
# Blender 5.2: RenderEngine.__subclasses__() may not exist
try:
    engines = sorted([e.identifier for e in bpy.types.RenderEngine.__subclasses__()])
except:
    engines = ["BLENDER_EEVEE", "BLENDER_CYCLES"]
    try:
        import _cycles
        engines.append("BLENDER_CYCLES")
    except:
        pass
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
# Blender 5.2: enum_items_static not always available, use known list
types = sorted([
    "CHILD_OF", "CLAMP_TO", "COPY_LOCATION", "COPY_ROTATION", "COPY_SCALE",
    "COPY_TRANSFORMS", "DAMPED_TRACK", "IK", "LIMIT_DISTANCE", "LIMIT_LOCATION",
    "LIMIT_ROTATION", "LIMIT_SCALE", "LOCK_TRACK", "MAINTAIN_VOLUME",
    "OBJECT_SOLVER", "PIVOT_TRACK", "SHRINKWRAP", "TRACK_TO", "TRANSFORM",
    "TRANSFORM_CACHE", "WORLD_TRANSFORM",
])
result = {"constraint_types": types, "count": len(types)}
'@
    },
    @{
        Id   = 15
        Desc = "Physics Categories"
        Code = @'
import bpy
# List physics types from enums/modifiers without creating objects
try:
    ff_types = sorted([t.identifier for t in bpy.types.Field.bl_rna.properties['type'].enum_items_static])
except:
    ff_types = ["FORCE", "WIND", "VORTEX", "MAGNET", "RHARBOR", "CHARGE", "LENNARDJENKINS", "TEXTURE", "HARMONIC", "TURBULENCE", "DRAG", "SMOKE_FLOW"]
try:
    fluid_types = sorted([t.identifier for t in bpy.types.FluidDomainSettings.bl_rna.properties.get('type', bpy.types.FluidDomainSettings.bl_rna.properties.get('domain_type', None)).enum_items_static]) if hasattr(bpy.types, 'FluidDomainSettings') else ["DOMAIN", "FLOW", "EFFECTOR"]
except:
    fluid_types = ["DOMAIN", "FLOW", "EFFECTOR"]
result = {
    "rigid_body_types": ["ACTIVE", "PASSIVE"],
    "fluid_types": fluid_types,
    "force_field_types": ff_types,
    "physics_modifiers": ["CLOTH", "SOFT_BODY", "FLUID", "COLLISION", "DYNAMIC_PAINT", "PARTICLE_SYSTEM", "SIMPLIFY", "MESH_SEQUENCE_CACHE"],
}
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
    try:
        info = bpy.utils.addon_info.get_addon_info(a.module)
        addons.append({
            "module": a.module,
            "name": info.name if info else a.module,
            "version": ".".join(str(v) for v in info.version) if info and info.version else "unknown",
        })
    except:
        addons.append({"module": a.module, "name": a.module, "version": "unknown"})
addons.sort(key=lambda x: x["module"])
result = {"addons": addons, "count": len(addons)}
'@
    },
    @{
        Id   = 18
        Desc = "Shader Node Classes"
        Code = @'
import bpy
shader = sorted([n.__name__ for n in bpy.types.ShaderNode.__subclasses__()])
compositor = sorted([n.__name__ for n in bpy.types.CompositorNode.__subclasses__()])
result = {
    "shader_nodes": shader, "shader_count": len(shader),
    "compositor_nodes": compositor, "compositor_count": len(compositor),
}
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
        Desc = "Operator Namespaces"
        Code = @'
import bpy
ns = sorted([n for n in dir(bpy.ops) if not n.startswith('_')])
result = {"operator_namespaces": ns, "count": len(ns)}
'@
    },
    @{
        Id   = 22
        Desc = "Grease Pencil"
        Code = @'
import bpy
gp_ops = [n for n in dir(bpy.ops.grease_pencil) if not n.startswith('_')] if hasattr(bpy.ops, 'grease_pencil') else []
gpencil_mods = [t.identifier for t in bpy.types.Modifier.bl_rna.properties['type'].enum_items_static if 'GPENCIL' in t.identifier or 'GREASE' in t.identifier]
result = {"gpencil_ops_count": len(gp_ops), "gpencil_ops_sample": gp_ops[:20], "gpencil_modifiers": gpencil_mods}
'@
    },
    @{
        Id   = 23
        Desc = "Blender Version & Paths"
        Code = @'
import bpy, sys, os
result = {
    "blender_version": bpy.app.version_string,
    "blender_version_tuple": list(bpy.app.version),
    "python_version": sys.version,
    "platform": sys.platform,
    "executable": sys.executable,
    "addons_path": bpy.utils.user_resource('SCRIPTS', path="addons"),
    "scripts_path": bpy.utils.user_resource('SCRIPTS'),
    "resources_path": bpy.utils.resource_path('LOCAL'),
}
'@
    }
)

# ── run probes ───────────────────────────────────────────────────────────────

Write-Host "Starting MCP server..." -ForegroundColor Yellow
$proc = Start-McpServer

# Phase 1: Fire all probes from main thread (just pipe writes, instant)
Write-Host "Firing $($probes.Count) probes..." -ForegroundColor Yellow
foreach ($p in $probes) {
    $msg = @{ jsonrpc = "2.0"; id = $p.Id; method = "tools/call"; params = @{ name = "execute_code"; arguments = @{ code = $p.Code } } }
    $json = $msg | ConvertTo-Json -Depth 10 -Compress
    $proc.StandardInput.WriteLine($json)
}
Write-Host "  All probes sent." -ForegroundColor DarkGray

# Phase 2: Close stdin, read responses with timeout
Write-Host "Closing stdin, reading responses..." -ForegroundColor Yellow
$proc.StandardInput.Close()

$lines = @()
$deadline = [DateTime]::UtcNow.AddSeconds(45)
$previousCount = 0
$stallCount = 0

Write-Host "  Reading (timeout 45s)..." -ForegroundColor Gray
while ([DateTime]::UtcNow -lt $deadline) {
    if ($proc.StandardOutput.EndOfStream) {
        Write-Host "  Stream ended." -ForegroundColor Gray
        break
    }
    $line = $proc.StandardOutput.ReadLine()
    if ($line) {
        $lines += $line
        $currentCount = $lines.Count
        if ($currentCount -gt $previousCount) {
            Write-Host "  Got response $currentCount..." -ForegroundColor DarkGray
            $previousCount = $currentCount
            $stallCount = 0
        }
    } else {
        $stallCount++
        if ($stallCount -gt 100) {
            Write-Host "  Stalled, stopping." -ForegroundColor Yellow
            break
        }
    }
}

$rawOutput = $lines -join "`n"
Write-Host "  Collected $($lines.Count) lines." -ForegroundColor $(if ($lines.Count -ge $probes.Count) { "Green" } else { "Yellow" })

# Kill server
if (-not $proc.HasExited) {
    try { $proc.Kill() } catch {}
    $proc.WaitForExit(2000) | Out-Null
}

# ── parse results ────────────────────────────────────────────────────────────

$results = @{}
$splitLines = $rawOutput -split "`n"
foreach ($line in $splitLines) {
    $line = $line.Trim()
    if (-not $line) { continue }
    try {
        $json = $line | ConvertFrom-Json
        if (($json.id -is [int] -or $json.id -is [long]) -and $json.id -ge 10 -and $null -ne $json.result.content) {
            $text = $json.result.content[0].text
            $results["$($json.id)"] = $text | ConvertFrom-Json
        }
    } catch {}
}

Write-Host "Parsed $($results.Count)/$($probes.Count) results" -ForegroundColor $(if ($results.Count -eq $probes.Count) { "Green" } else { "Yellow" })

# ── write markdown ───────────────────────────────────────────────────────────

$sb = [System.Text.StringBuilder]::new()
$codeFence = [string][char]0x60 * 3

[void]$sb.AppendLine("# Blender MCP Capability Probe")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("**Date:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')")
[void]$sb.AppendLine("**Blender MCP Server:** $McpPath")
[void]$sb.AppendLine("")

foreach ($p in $probes) {
    $data = $results["$($p.Id)"]
    [void]$sb.AppendLine("## $($p.Desc)")
    [void]$sb.AppendLine("")

    if ($data) {
        $json = $data | ConvertTo-Json -Depth 10
        [void]$sb.AppendLine("${codeFence}json")
        [void]$sb.AppendLine($json)
        [void]$sb.AppendLine($codeFence)
    }
    else {
        [void]$sb.AppendLine("_No result returned (timeout or error)._")
    }
    [void]$sb.AppendLine("")
}

# ── summary ──────────────────────────────────────────────────────────────────

[void]$sb.AppendLine("## Summary")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("| Probe | Status | Key Info |")
[void]$sb.AppendLine("|-------|--------|----------|")

foreach ($p in $probes) {
    $data = $results["$($p.Id)"]
    if ($data) {
        # Try to extract a useful summary line
        $summary = ""
        $props = $data.PSObject.Properties
        foreach ($prop in $props) {
            if ($prop.Value -is [array] -and $prop.Name -like "*count*") {
                $summary = "$($prop.Name): $($prop.Value)"
                break
            }
            elseif ($prop.Value -is [array]) {
                $summary = "$($prop.Name): $($prop.Value.Count) items"
                break
            }
            elseif ($prop.Name -eq "current" -or $prop.Name -eq "engine") {
                $summary = "$($prop.Name): $($prop.Value)"
            }
        }
        [void]$sb.AppendLine("| $($p.Desc) | OK | $summary |")
    }
    else {
        [void]$sb.AppendLine("| $($p.Desc) | TIMEOUT | - |")
    }
}

[void]$sb.AppendLine("")
[void]$sb.AppendLine("---")
[void]$sb.AppendLine("*Generated by probe-blender.ps1*")

$sb.ToString() | Out-File -FilePath $OutputPath -Encoding UTF8
Write-Host ""
Write-Host "Results written to: $OutputPath" -ForegroundColor Green
