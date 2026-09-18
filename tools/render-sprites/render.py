"""Headless Blender pipeline for Botropolis' sprites.

    blender -b --python tools/render-sprites/render.py -- scene --out out.png
    blender -b --python tools/render-sprites/render.py -- atlas --out pkg/assets/kits
    blender -b --python tools/render-sprites/render.py -- --help

The `scene` mode composes one district from Kenney's City Kits — Commercial
for the buildings, Roads for the avenues, Industrial for the plant — and
renders it once from the isometric heading, so the look can be judged
beside the current map before any atlas is cut.

The `atlas` mode renders every piece the map uses at four headings and
each zoom level into atlas pages with a JSON manifest: for each sprite
its page, its rectangle and where the piece's ground origin lands, so
the renderer can set it on a cell.
"""

import argparse
import json
import math
import os
import sys

import bpy
import numpy as np
from bpy_extras.object_utils import world_to_camera_view
from mathutils import Vector

KITS = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "kits")

# The map's 2:1 diamond: the camera looks down at atan(1/2) above the
# ground, turned 45 degrees, so a tile's top face projects 2:1.
ISO_TILT = math.degrees(math.atan2(1, 0.5))
ISO_TURN = 45.0

# One kit unit is one map cell; the map draws a cell as a diamond
# TILE_PX wide at zoom 1 (city.IsoTileWidth), and a unit's diagonal lies
# flat across that diamond.
TILE_PX = 132.0
PX_PER_UNIT = TILE_PX / math.sqrt(2)
PAGE = 2048
PAGE_BUDGET = 8
PAD = 2

# Every piece the map draws, by kit: the atlas is cut from this list.
PIECES = {
    "city-kit-commercial": [f"building-{c}" for c in "abcdefghijklmn"] + [f"building-skyscraper-{c}" for c in "abcde"],
    "city-kit-industrial": ["building-a", "building-b", "building-e", "building-h", "building-k", "chimney-large", "chimney-medium",
                            "detail-tank-large", "water-tower", "windmill", "solar-panel-landscape-group",
                            "shipping-container-a", "shipping-container-b", "shipping-container-c"],
    "city-kit-roads": ["road-straight", "road-bend", "road-curve", "road-crossroad", "road-intersection", "road-end", "road-square",
                       "light-square", "light-curved", "electricity-pole", "electricity-wires", "traffic-light", "construction-cone"],
    "city-kit-suburban": ["tree-large", "tree-small"],
    "car-kit": ["sedan", "van", "taxi", "suv", "hatchback-sports", "truck", "delivery"],
    "space-kit": ["rover"],
    "train-kit": ["train-diesel-a", "train-diesel-b", "train-diesel-c",
                  "train-carriage-container-red", "train-carriage-container-blue", "train-carriage-container-green",
                  "train-carriage-box", "train-carriage-tank"],
    "watercraft-kit": ["boat-tug-a", "boat-tug-b", "boat-row-small"],
    "botropolis": ["drone"],
}

# Kits that are not modelled at one unit per cell are scaled on import:
# the Car Kit is in metres, a sedan 2.5 long, and a car on the map is a
# third of a cell; the Train Kit's wagons are 2.7 long and the Watercraft
# Kit's tug 3.5, and a wagon or a barge on the map is two thirds of a cell.
SCALE = {"car-kit": 0.12, "train-kit": 0.25, "watercraft-kit": 0.25}

GRASS = (0.22, 0.33, 0.17, 1)
PLAZA = (0.54, 0.53, 0.49, 1)
CONCRETE = (0.45, 0.45, 0.47, 1)


def parse_args():
    argv = sys.argv[sys.argv.index("--") + 1 :] if "--" in sys.argv else []
    p = argparse.ArgumentParser(prog="render.py", description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("mode", choices=["scene", "atlas"], help="what to render")
    p.add_argument("--out", required=True, help="PNG to write (scene) or directory for the atlases (atlas)")
    p.add_argument("--zooms", default="1,2", help="zoom levels to cut atlases for (atlas)")
    p.add_argument("--only", default="", help="render only pieces whose name contains this (atlas, for trying things)")
    p.add_argument("--kits", default=KITS, help="where the Kenney kits are (tools/kits)")
    p.add_argument("--width", type=int, default=1600)
    p.add_argument("--height", type=int, default=1000)
    p.add_argument("--scale", type=float, default=13.0, help="orthographic width in tiles")
    return p.parse_args(argv)


def reset():
    bpy.ops.wm.read_factory_settings(use_empty=True)


ACCENT = (0.91, 0.63, 0.24, 1)
SLATE = (0.16, 0.18, 0.22, 1)
OFFWHITE = (0.85, 0.86, 0.88, 1)


def solid(name, colour, emission=0.0):
    mat = bpy.data.materials.new(name)
    mat.use_nodes = True
    bsdf = mat.node_tree.nodes["Principled BSDF"]
    bsdf.inputs["Base Color"].default_value = colour
    bsdf.inputs["Roughness"].default_value = 0.7
    if emission:
        bsdf.inputs["Emission Color"].default_value = colour
        bsdf.inputs["Emission Strength"].default_value = emission
    return mat


def drone(at):
    """Our own drone: a slate body, two off-white rotors, one accent
    light — the subagent in flight. Modelled here so it shares the kits'
    palette and needs no third-party asset."""
    root = bpy.data.objects.new("botropolis/drone", None)
    bpy.context.scene.collection.objects.link(root)
    parts = []
    bpy.ops.mesh.primitive_cylinder_add(radius=0.09, depth=0.07, location=(0, 0, 0.32))
    body = bpy.context.active_object
    body.data.materials.append(solid("drone-body", SLATE))
    parts.append(body)
    bpy.ops.mesh.primitive_cube_add(size=1, location=(0, 0, 0.35))
    arm = bpy.context.active_object
    arm.scale = (0.34, 0.03, 0.015)
    arm.data.materials.append(solid("drone-arm", SLATE))
    parts.append(arm)
    for x in (-0.16, 0.16):
        bpy.ops.mesh.primitive_cylinder_add(radius=0.1, depth=0.012, location=(x, 0, 0.37))
        rotor = bpy.context.active_object
        rotor.data.materials.append(solid("drone-rotor", OFFWHITE))
        parts.append(rotor)
    bpy.ops.mesh.primitive_uv_sphere_add(radius=0.035, location=(0, 0, 0.27))
    lamp = bpy.context.active_object
    lamp.data.materials.append(solid("drone-light", ACCENT, emission=3.0))
    parts.append(lamp)
    for o in parts:
        o.parent = root
    root.location = Vector((at[0], at[1], 0))
    bpy.context.view_layer.update()
    return root


def model_path(kits, kit, name):
    for folder in ("GLB format", "GLTF format"):
        path = os.path.join(kits, kit, "Models", folder, name + ".glb")
        if os.path.exists(path):
            return path
    raise FileNotFoundError(f"{kit}/{name}.glb is not in tools/kits; run make kits")


def piece(kits, kit, name, at, turn=0.0):
    """Import one kit piece at a tile position, turned in degrees about z;
    the botropolis kit is modelled here rather than imported."""
    if kit == "botropolis":
        if name != "drone":
            raise FileNotFoundError(f"botropolis/{name} is not a piece we model")
        return drone(at)
    before = set(bpy.context.scene.objects)
    bpy.ops.import_scene.gltf(filepath=model_path(kits, kit, name))
    new = [o for o in bpy.context.scene.objects if o not in before]
    root = bpy.data.objects.new(f"{kit}/{name}", None)
    bpy.context.scene.collection.objects.link(root)
    for o in new:
        if o.parent is None:
            o.parent = root
    root.location = Vector((at[0], at[1], at[2] if len(at) > 2 else 0))
    root.rotation_euler = (0, 0, math.radians(turn))
    k = SCALE.get(kit, 1.0)
    root.scale = (k, k, k)
    bpy.context.view_layer.update()
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


def remove(root):
    """Delete an imported piece and everything under it."""
    for o in list(root.children_recursive) + [root]:
        bpy.data.objects.remove(o, do_unlink=True)
    for block in (bpy.data.meshes, bpy.data.materials, bpy.data.images):
        for datum in list(block):
            if datum.users == 0:
                block.remove(datum)


def bounds(root):
    lo = Vector((1e9, 1e9, 1e9))
    hi = Vector((-1e9, -1e9, -1e9))
    for o in root.children_recursive:
        if o.type != "MESH":
            continue
        for c in o.bound_box:
            w = o.matrix_world @ Vector(c)
            lo = Vector(map(min, lo, w))
            hi = Vector(map(max, hi, w))
    return lo, hi


def frame_piece(root, heading, zoom):
    """Point the camera at a piece from a heading and size the frame to
    hold it: returns the resolution and where the ground origin lands."""
    scene = bpy.context.scene
    cam = scene.camera
    cam.rotation_euler = (math.radians(ISO_TILT), 0, math.radians(ISO_TURN + heading))
    lo, hi = bounds(root)
    centre = (lo + hi) / 2
    back = cam.rotation_euler.to_matrix() @ Vector((0, 0, 60))
    cam.location = centre + back
    bpy.context.view_layer.update()
    corners = [Vector((x, y, z)) for x in (lo.x, hi.x) for y in (lo.y, hi.y) for z in (lo.z, hi.z)]
    # Ortho scale is the frame's width in camera units; find the extent
    # of the piece in camera space and add a margin.
    view = cam.rotation_euler.to_matrix().inverted()
    xs = [(view @ (c - cam.location)).x for c in corners]
    ys = [(view @ (c - cam.location)).y for c in corners]
    margin = 0.15
    w = max(xs) - min(xs) + 2 * margin
    h = max(ys) - min(ys) + 2 * margin
    size = max(w, h)
    cam.data.ortho_scale = size
    px = int(math.ceil(size * PX_PER_UNIT * zoom))
    scene.render.resolution_x = px
    scene.render.resolution_y = px
    bpy.context.view_layer.update()
    ndc = world_to_camera_view(scene, cam, Vector((0, 0, 0)))
    return px, (ndc.x * px, (1 - ndc.y) * px)


def render_pixels(path):
    """Render the frame to path and return it as an HxWx4 uint8 array."""
    scene = bpy.context.scene
    scene.render.filepath = path
    bpy.ops.render.render(write_still=True)
    img = bpy.data.images.load(path)
    w, h = img.size
    buf = np.empty(w * h * 4, dtype=np.float32)
    img.pixels.foreach_get(buf)
    bpy.data.images.remove(img)
    rgba = (buf.reshape(h, w, 4)[::-1] * 255 + 0.5).astype(np.uint8)
    return rgba


def crop(rgba):
    alpha = rgba[:, :, 3]
    rows = np.where(alpha.any(axis=1))[0]
    cols = np.where(alpha.any(axis=0))[0]
    if len(rows) == 0:
        return rgba[:1, :1], (0, 0)
    y0, y1 = rows[0], rows[-1] + 1
    x0, x1 = cols[0], cols[-1] + 1
    return rgba[y0:y1, x0:x1], (x0, y0)


class Pages:
    """A shelf packer over PAGE-square pages."""

    def __init__(self):
        self.pages = []
        self.shelf_y = 0
        self.shelf_h = 0
        self.x = 0

    def new_page(self):
        self.pages.append(np.zeros((PAGE, PAGE, 4), dtype=np.uint8))
        self.shelf_y = self.shelf_h = self.x = 0

    def put(self, rgba):
        h, w = rgba.shape[:2]
        if w + 2 * PAD > PAGE or h + 2 * PAD > PAGE:
            raise ValueError(f"a sprite of {w}x{h} does not fit a {PAGE} page; lower the zoom or split the piece")
        if not self.pages:
            self.new_page()
        if self.x + w + PAD > PAGE:
            self.shelf_y += self.shelf_h + PAD
            self.shelf_h = 0
            self.x = 0
        if self.shelf_y + h + PAD > PAGE:
            self.new_page()
        page = len(self.pages) - 1
        x, y = self.x + PAD, self.shelf_y + PAD
        self.pages[page][y : y + h, x : x + w] = rgba
        self.x += w + PAD
        self.shelf_h = max(self.shelf_h, h)
        return page, x, y

    def save(self, out, stem):
        names = []
        for i, page in enumerate(self.pages):
            name = f"{stem}-{i}.png"
            img = bpy.data.images.new(name, PAGE, PAGE, alpha=True)
            img.pixels.foreach_set((page[::-1].astype(np.float32) / 255).ravel())
            img.filepath_raw = os.path.join(out, name)
            img.file_format = "PNG"
            img.save()
            bpy.data.images.remove(img)
            names.append(name)
        return names


def atlas(args):
    """Cut the atlases: every piece, four headings, each zoom level."""
    out = os.path.abspath(args.out)
    os.makedirs(out, exist_ok=True)
    tmp = os.path.join(out, ".render.png")
    zooms = [float(z) for z in args.zooms.split(",") if z]
    reset()
    settings()
    bpy.context.scene.render.film_transparent = True
    light()
    camera((0, 0, 0), 4, 64, 64)
    for zoom in zooms:
        pages = Pages()
        sprites = {}
        count = 0
        for kit, names in PIECES.items():
            for name in names:
                if args.only and args.only not in name:
                    continue
                root = piece(args.kits, kit, name, (0, 0))
                key = f"{kit}/{name}"
                sprites[key] = {}
                for heading in (0, 90, 180, 270):
                    px, (ax, ay) = frame_piece(root, heading, zoom)
                    rgba = render_pixels(tmp)
                    cut, (x0, y0) = crop(rgba)
                    page, x, y = pages.put(cut)
                    sprites[key][str(heading)] = {
                        "page": page, "x": int(x), "y": int(y), "w": int(cut.shape[1]), "h": int(cut.shape[0]),
                        "ax": round(float(ax - x0), 2), "ay": round(float(ay - y0), 2),
                    }
                    count += 1
                remove(root)
        stem = f"kits-z{zoom:g}"
        names = pages.save(out, stem)
        if len(names) > PAGE_BUDGET:
            print(f"BUDGET zoom {zoom:g} needs {len(names)} pages; the budget is {PAGE_BUDGET}", file=sys.stderr)
            sys.exit(1)
        manifest = {"zoom": zoom, "tile": TILE_PX * zoom, "page": PAGE, "pages": names, "sprites": sprites}
        with open(os.path.join(out, stem + ".json"), "w") as f:
            json.dump(manifest, f, indent=1, sort_keys=True)
            f.write("\n")
        print(f"WROTE {stem}: {count} sprites on {len(names)} page(s)")
    if os.path.exists(tmp):
        os.remove(tmp)


def main():
    args = parse_args()
    if args.mode == "atlas":
        atlas(args)
        return
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
