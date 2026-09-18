"""Headless Blender pipeline for Botropolis' sprites.

    blender -b --python tools/render-sprites/render.py -- scene --out out.png
    blender -b --python tools/render-sprites/render.py -- --help

The `scene` mode composes one district from Kenney's City Kits — Commercial
for the buildings, Roads for the avenues, Industrial for the plant — and
renders it once from the isometric heading, so the look can be judged
beside the current map before any atlas is cut.
"""

import argparse
import math
import os
import sys

import bpy
from mathutils import Vector

KITS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "kits")

# The map's 2:1 diamond: the camera looks down at atan(1/2) above the
# ground, turned 45 degrees, so a tile's top face projects 2:1.
ISO_TILT = math.degrees(math.atan2(1, 0.5))
ISO_TURN = 45.0

GRASS = (0.22, 0.33, 0.17, 1)
PLAZA = (0.54, 0.53, 0.49, 1)
CONCRETE = (0.45, 0.45, 0.47, 1)


def parse_args():
    argv = sys.argv[sys.argv.index("--") + 1 :] if "--" in sys.argv else []
    p = argparse.ArgumentParser(prog="render.py", description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("mode", choices=["scene"], help="what to render")
    p.add_argument("--out", required=True, help="PNG to write")
    p.add_argument("--kits", default=KITS, help="where the Kenney kits are (tools/kits)")
    p.add_argument("--width", type=int, default=1600)
    p.add_argument("--height", type=int, default=1000)
    p.add_argument("--scale", type=float, default=13.0, help="orthographic width in tiles")
    return p.parse_args(argv)


def reset():
    bpy.ops.wm.read_factory_settings(use_empty=True)


def piece(kits, kit, name, at, turn=0.0):
    """Import one kit piece at a tile position, turned in degrees about z."""
    before = set(bpy.context.scene.objects)
    bpy.ops.import_scene.gltf(filepath=os.path.join(kits, kit, "Models", "GLB format", name + ".glb"))
    new = [o for o in bpy.context.scene.objects if o not in before]
    root = bpy.data.objects.new(f"{kit}/{name}", None)
    bpy.context.scene.collection.objects.link(root)
    for o in new:
        if o.parent is None:
            o.parent = root
    root.location = Vector((at[0], at[1], at[2] if len(at) > 2 else 0))
    root.rotation_euler = (0, 0, math.radians(turn))
    return root


def flat(name, size, at, colour):
    """A coloured plane: ground, plaza floor, concrete apron."""
    bpy.ops.mesh.primitive_plane_add(size=1, location=(at[0], at[1], at[2] if len(at) > 2 else 0))
    o = bpy.context.active_object
    o.name = name
    o.scale = (size[0], size[1], 1)
    mat = bpy.data.materials.new(name)
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes["Principled BSDF"]
    bsdf.inputs["Base Color"].default_value = colour
    bsdf.inputs["Roughness"].default_value = 1.0
    o.data.materials.append(mat)
    return o


def light():
    sun = bpy.data.lights.new("sun", "SUN")
    sun.energy = 4.0
    sun.angle = math.radians(3)
    o = bpy.data.objects.new("sun", sun)
    bpy.context.scene.collection.objects.link(o)
    # One light direction for every sprite: high, from the front-left.
    o.rotation_euler = (math.radians(48), math.radians(12), math.radians(-35))
    world = bpy.data.worlds.new("world")
    world.use_nodes = True
    bg = world.node_tree.nodes["Background"]
    bg.inputs["Color"].default_value = (0.7, 0.78, 0.9, 1)
    bg.inputs["Strength"].default_value = 0.55
    bpy.context.scene.world = world


def camera(centre, scale, width, height):
    cam = bpy.data.cameras.new("iso")
    cam.type = "ORTHO"
    cam.ortho_scale = scale
    cam.clip_end = 500
    o = bpy.data.objects.new("iso", cam)
    bpy.context.scene.collection.objects.link(o)
    o.rotation_euler = (math.radians(ISO_TILT), 0, math.radians(ISO_TURN))
    # Back the camera off along its own view axis so everything is in front.
    back = o.rotation_euler.to_matrix() @ Vector((0, 0, 60))
    o.location = Vector(centre) + back
    bpy.context.scene.camera = o
    r = bpy.context.scene.render
    r.resolution_x, r.resolution_y = width, height
    r.resolution_percentage = 100


def settings():
    scene = bpy.context.scene
    scene.render.engine = "BLENDER_EEVEE"
    scene.render.film_transparent = False
    scene.render.image_settings.file_format = "PNG"
    scene.render.image_settings.color_mode = "RGBA"
    scene.view_settings.view_transform = "Standard"
    scene.view_settings.look = "None"
    eevee = scene.eevee
    for attr, value in (("taa_render_samples", 32), ("use_gtao", True), ("use_shadows", True), ("use_soft_shadows", True), ("shadow_cube_size", "2048"), ("shadow_cascade_size", "4096")):
        if hasattr(eevee, attr):
            try:
                setattr(eevee, attr, value)
            except Exception:
                pass


def district(kits):
    """One live district on its block, the plant on the plaza next door,
    avenues round both, lamps at the crossings."""
    flat("ground", (40, 40), (3.5, 1, -0.02), GRASS)
    # Blocks: the district at x 0..4, y 0..3; the plaza at x 5..8.
    flat("block", (4, 3), (2, 1.5, 0.0), PLAZA)
    flat("plaza", (3, 3), (6.5, 1.5, 0.0), CONCRETE)
    # Roads: two rows and three columns of avenue round the blocks.
    for x in range(0, 4):
        piece(kits, "city-kit-roads", "road-straight", (x + 0.5, -0.5), 0)
        piece(kits, "city-kit-roads", "road-straight", (x + 0.5, 3.5), 0)
    for x in range(5, 8):
        piece(kits, "city-kit-roads", "road-straight", (x + 0.5, -0.5), 0)
        piece(kits, "city-kit-roads", "road-straight", (x + 0.5, 3.5), 0)
    for y in range(0, 3):
        for x in (-0.5, 4.5, 8.5):
            piece(kits, "city-kit-roads", "road-straight", (x, y + 0.5), 90)
    piece(kits, "city-kit-roads", "road-bend", (-0.5, -0.5), 180)
    piece(kits, "city-kit-roads", "road-bend", (8.5, -0.5), 270)
    piece(kits, "city-kit-roads", "road-bend", (8.5, 3.5), 0)
    piece(kits, "city-kit-roads", "road-bend", (-0.5, 3.5), 90)
    piece(kits, "city-kit-roads", "road-intersection", (4.5, -0.5), 90)
    piece(kits, "city-kit-roads", "road-intersection", (4.5, 3.5), 270)
    # Lamps at the crossings.
    for at, turn in (((-0.15, -0.15), 0), ((4.15, -0.15), 90), ((8.15, -0.15), 90), ((-0.15, 3.15), 270), ((4.15, 3.15), 180), ((8.15, 3.15), 180)):
        piece(kits, "city-kit-roads", "light-square", at, turn)
    # Sessions: six buildings on the block, one per session.
    names = ["building-a", "building-b", "building-c", "building-d", "building-e", "building-f"]
    i = 0
    for row in range(2):
        for col in range(3):
            piece(kits, "city-kit-commercial", names[i], (0.7 + col * 1.3, 0.75 + row * 1.5), 0)
            i += 1
    # The plant: an industrial hall, its stack, a tank and the water tower.
    piece(kits, "city-kit-industrial", "building-a", (6.3, 1.0), 0)
    piece(kits, "city-kit-industrial", "chimney-large", (7.4, 2.3), 0)
    piece(kits, "city-kit-industrial", "detail-tank-large", (5.6, 2.4), 0)
    piece(kits, "city-kit-industrial", "water-tower", (7.5, 0.6), 0)


def main():
    args = parse_args()
    reset()
    settings()
    light()
    district(args.kits)
    camera((4.0, 1.5, 0.4), args.scale, args.width, args.height)
    bpy.context.scene.render.filepath = os.path.abspath(args.out)
    bpy.ops.render.render(write_still=True)
    print("WROTE", bpy.context.scene.render.filepath)


if __name__ == "__main__":
    main()
